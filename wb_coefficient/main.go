package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	configFile          = "config.json"
	minimalPauseRequest = 15
	appNameInRedis      = "public_bot"
	EmojiInbox          = "📥"
	EmojiSuccess        = "✅"
	EmojiProcessing     = "⚡"
	EmojiWarning        = "⚠️"
	EmojiClock          = "🕒"
	EmojiStats          = "📊"
	EmojiError          = "❌"
	EmojiClient         = "👤"
	EmojiLoop           = "🔄"
	EmojiTelegram       = "📨"
)

var (
	ctx          context.Context
	cancel       context.CancelFunc
	shutdownChan = make(chan os.Signal, 1)
	logging      func(string, ...any)
	logs         strings.Builder
	logsCapacity = 2 * 1024
	logMutex     sync.Mutex
	redisClient  *redis.Client
	redisConfig  *RedisConfig
	appConfig    *AppConfig
	httpClient   = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DisableCompression:    false,
			ResponseHeaderTimeout: 8 * time.Second,
			TLSHandshakeTimeout:   3 * time.Second,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxConnsPerHost:       2,
		},
	}
)

func main() {

	// объявление контекста завершения приложения
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// ловим определённые сигналы для завершения приложения
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// переменная для логирования времени работы приложения
	startGlobalTime := time.Now()

	// объявляем логирование и размер логов
	logs.Grow(logsCapacity)
	// функция логирования
	logging = func(data string, args ...any) {
		logMutex.Lock()
		defer logMutex.Unlock()
		timeStamp := time.Now().Format("15:04:05.000")
		// сброс логов в StdOut
		fmt.Fprintf(&logs, "[%s] ", timeStamp)
		if len(args) > 0 {
			fmt.Fprintf(&logs, data, args...)
		} else {
			logs.WriteString(data)
		}
		logs.WriteByte('\n')
	}

	// фоновая горутина, которая ловит сигнал завершения приложения
	go func() {
		sig := <-shutdownChan
		logging("%s получен сигнал завершения: %v", EmojiWarning, sig)
		cancel()
		time.Sleep(2 * time.Second)
	}()

	logging("🚀 запускаемся...")

	// горутина для ловли паник от падения приложения
	defer func() {
		if r := recover(); r != nil {
			log.Printf("%s паника в основном потоке: %v", EmojiWarning, r)
			debug.PrintStack()
		}
	}()

	// сброс логов при завершении приложения
	defer func() {
		if logs.Len() == 0 {
			return
		}
		if _, err := fmt.Print(logs.String()); err != nil {
			log.Printf("%s возникла ошибка при записи логов: %v", EmojiWarning, err)
		}
	}()

	// горутина с логированием времени работы приложения с момента запуска демона
	defer func() {
		var (
			msg      strings.Builder
			workTime = int(time.Since(startGlobalTime).Seconds())
		)
		msg.Grow(64)
		if days := workTime / 86400; days > 0 {
			fmt.Fprintf(&msg, "%d дней ", days)
		}
		if hours := (workTime % 86400) / 3600; hours > 0 {
			fmt.Fprintf(&msg, "%d часов ", hours)
		}
		if minutes := (workTime % 3600) / 60; minutes > 0 {
			fmt.Fprintf(&msg, "%d минут ", minutes)
		}
		seconds := workTime % 60
		fmt.Fprintf(&msg, "%d секунд", seconds)
		logging("%s приложение завершено, общее время работы: %s",
			EmojiSuccess, msg.String())
	}()

	// создание Редис клиента
	redisConfig = &RedisConfig{
		Addr:     os.Getenv("redisAddr"),
		Password: os.Getenv("redisPassword"),
		DB:       0,
		TimeOut:  3 * time.Second}

	var err error

	// стартовые проверки
	// подключение к Редис
	// загрузка с перезаписью уже имеющихся значений
	// проверка на загруженный конфиг в Редис и в приложение
	logging("📡 подключаемся к Redis...")
	if redisClient, err = checkRedisConnection(); err != nil {
		logging("%s ошибка запуска возникла при проверке подключения к Redis с полученными аргументами запуска приложения: %v", EmojiError, err)
		return
	}
	logging("📋 загружаем конфигурацию...")
	if err := loadConfigFromJson(); err != nil {
		logging("%s ошибка загрузки конфигурации: %v", EmojiError, err)
		return
	}
	if appConfig == nil {
		logging("%s КОНФИГ НЕ ЗАГРУЖЕН! appConfig is nil", EmojiError)
		return
	}
	// проверка флага на работу приложения
	if !appConfig.Working {
		logging("%s приложение на паузе параметр [working] в config.json", EmojiWarning)
		return
	}

	// проверки закончены, старт приложения
	logging("%s приложение успешно запущено", EmojiSuccess)

	// создаём срез необходимого размера для аллокации в памяти один раз
	var data = make([]Response, 0, 1024)

	// вечный цикл (приложение - демон)
	for c := 0; ; c++ {
		// обнуление среза с данными после каждой итерации
		data = data[:0]
		// кэшируем даты, чтобы постоянно их не преобразовывать
		var mapDate = make(map[string]string)
		// ловим команду на завершение приложения
		if ctx.Err() != nil {
			logging("%s получена команда остановки приложения", EmojiWarning)
			time.Sleep(100 * time.Millisecond)
			return
		}

		// если вдруг приложение стало на паузу, надо сообщить админу и просто держать паузу
		if !appConfig.Working {
			messageForAdmin := fmt.Sprintf("%s приложение на паузе, ждем 300 секунд", EmojiWarning)
			logging("%s", messageForAdmin)
			if c%5 == 0 {
				if err := sendTextMessage(messageForAdmin, appConfig.Admin, 0); err != nil {
					logging("%v", err)
				}
			}

			// ловим сигнал на завершение работы
			for range 300 {
				if ctx.Err() != nil {
					logging("%s получена команда остановки приложения", EmojiWarning)
					time.Sleep(1 * time.Second)
					return
				}
				time.Sleep(1 * time.Second)
			}
			continue
		}

		// переменные для логирования времени работы
		// расчёт паузы, минимальная пауза 10 сек, мы вычисляем большее значение из двух
		var (
			startIterationTime = time.Now()
			pauseInIteration   = max(minimalPauseRequest, appConfig.PauseIteration)
		)

		// получаем сырую информацию от api WB по коэффициентам приёмки
		if err := getCoefWarehouses(&data); err != nil {
			logging("%s ошибка при получении коэффициентов:\n%v", EmojiWarning, err)
			continue
		}
		logging("%s получено сырых данных: %d, capacity: %d", EmojiInbox, len(data), cap(data))

		// чистим данные от -1
		if err := clearData(&data, mapDate); err != nil {
			logging("%s ошибка очистки данных от КФ [-1]:\n%v", EmojiWarning, err)
			continue
		}

		// цикл обработки клиентов
		// отправка каждому клиенту своих данных в тележку
		for _, client := range appConfig.AllActiveClients {

			// пропуск итерациии, если по какой-то причине получился пустой срез
			// здесь, потому-что нужно соблюдать паузу между итерациями
			if len(data) == 0 {
				continue
			}

			// пропуск клиента у которого нет складов (новичок или кто удалил склады из конфига)
			if len(appConfig.Clients[client].BoxData)+len(appConfig.Clients[client].MonoData) == 0 {
				logging("%s пропуск клиента [%s], нет складов в конфигурации", EmojiWarning, client)
				continue
			}

			// пропуск клиента, если его конфиг не загружен по какой-то причине
			if client != appConfig.Admin {
				if _, ok := appConfig.Clients[client]; !ok {
					logging("пропуск клиента %s, его нет в конфигурации", client)
					continue
				}
			}

			// логирование времени отработки каждого клиента
			var (
				startWorkTimeClient = time.Now()
			)

			// если у клиента есть пауза по api от ТГ или ВБ, то минусуем и пропускаем клиента
			if appConfig.Clients[client].Pause > 0 {
				logging("%s у клиента [%s] пауза %d сек",
					EmojiWarning, client, appConfig.Clients[client].Pause)
				updatedClient := appConfig.Clients[client]
				if updatedClient.Pause > pauseInIteration {
					updatedClient.Pause -= pauseInIteration
				} else {
					updatedClient.Pause = 0
				}
				// сохраняем новую паузу за вычетом времени затраченного на итерацию
				appConfig.Clients[client] = updatedClient
				continue
			}

			// отправка очищенных данных в функцию формирования сообщения с последующей отправкой клиенту в чат
			if err := prepareMessages(data, mapDate, client); err != nil {
				logging("%v", err)
			}

			// финишное логирование каждого клиента и времени на его обработку
			logging("%s обработка [%s]: %.3f сек",
				EmojiProcessing, client, time.Since(startWorkTimeClient).Seconds())
		}

		// перезагрузка конфигурации
		// проверка флаг-ключа загруженного конфига в Редис
		reloadConfig, err := checkExistsKeyInRedis(appNameInRedis)
		if err != nil {
			logging("%s %v", EmojiWarning, err)
		}
		// если конфига нет в Редис или ttl флаг-ключа меньше текущей паузы -> загружаем конфиг
		ttlKey, err := checkTTLRedisKey(appNameInRedis)
		if err != nil {
			logging("%s %v", EmojiWarning, err)
		} else if !reloadConfig || ttlKey <= pauseInIteration {
			if err := loadConfigFromJson(); err != nil {
				logging("%s %v", EmojiWarning, err)
			}
		}

		// перезагрузка списка всех складов ВБ
		// если ключа "warehouse_list" нет в Редис
		// если такой ключ есть, то загружать заново не надо
		reloadListWHID, err := checkExistsKeyInRedis("warehouse_list")
		if err != nil {
			logging("%v", err)
		}
		if !reloadListWHID {
			// делаем мапу со складами по ключу "ID склада"
			// чтобы в дальнейшем искать по этим данным
			var listWarehouseID = make(map[int64]string)
			if err := getListWarehouseWB(&listWarehouseID); err != nil {
				logging("%s %v", EmojiWarning, err)
			} else {
				logging("%s получено %d складов для записи в Редис",
					EmojiSuccess, len(listWarehouseID))
				// формирование ключа для Редис
				var sb strings.Builder
				sb.Grow(20)
				for k, v := range listWarehouseID {
					sb.Reset()
					sb.WriteString("warehouse_")
					sb.WriteString(strconv.FormatInt(k, 10))
					// непосредственно загрузка данных в Редис
					// загрузка по одному складу в Редис
					if err := setStringRedis(sb.String(), v); err != nil {
						logging("%s при сохранении списка сладов в Редис, что-то пошло не так: %v",
							EmojiWarning, err)
					}
				}
				// формируем ключ-флаг для проверок на необходимость повторных загрузок
				if err := setStringRedis("warehouse_list", "OK"); err != nil {
					logging("%s при записи списка складов в Редис произошла ошибка: %v",
						EmojiWarning, err)
				}
			}
		}

		// расчёт паузы для итерации между запросами к api-WB
		sleep := time.Duration(pauseInIteration)*time.Second - time.Since(startIterationTime)

		// простое логирование между итерациями
		logging("%s время работы: %.1f сек, остаток паузы: %d сек",
			EmojiClock,
			time.Since(startIterationTime).Seconds(),
			int(sleep.Seconds()),
		)
		logging("%s всего отправлено [%d] сообщений в Телеграмм",
			EmojiTelegram, appConfig.AllCountSendMessages)
		appConfig.AllCountSendMessages = 0

		// непосредственно сброс логов в StdOut
		if logs.Len() > 0 {
			fmt.Print(logs.String())
			logs.Reset()
		}

		// на этом этапе просто спим между запросами к api-WB
		if sleep <= 0 {
			// минимальная пауза, на всякий случай
			time.Sleep(1 * time.Second)
		} else {

			// в этой части разбиваем на целые секунды
			// для проверки контекста завершения приложения
			seconds := int(sleep.Seconds())
			remainder := sleep - time.Duration(seconds)*time.Second
			for range seconds {
				if ctx.Err() != nil {
					logging("%s получен сигнал завершения приложения", EmojiWarning)
					return
				}
				time.Sleep(1 * time.Second)
			}
			// досыпаем остаток от целых секунд
			if remainder > 0 {
				time.Sleep(remainder)
			}
		}
	}
}
