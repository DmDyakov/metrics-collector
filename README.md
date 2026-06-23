# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Запустить сервер локально
go run cmd/server/main.go

## Запустить агент локально
go run cmd/agent/main.go

## Красивый вывод тестов
### Установка gotestfmt
go install github.com/gotesttools/gotestfmt/v2/cmd/gotestfmt@latest
### Команда для вывода
go test -v -json ./... 2>&1 | gotestfmt

## Генерация моков
### Установить mockgen
go install go.uber.org/mock/mockgen@latest

### После установки проверьте:
mockgen --version

### Обновить сгенерированные моки
go generate ./...

## Инкремент 17 Сравнение профилей после оптимизации
File: server_result
Build ID: C:\Dev\Go\metrics-collector\profiles\server_result2026-06-23 08:00:01.0274394 +0300 MSK
Type: inuse_space
Time: 2026-06-22 13:46:38 MSK
Showing nodes accounting for 5722.89kB, 65.02% of 8802.19kB total
      flat  flat%   sum%        cum   cum%
 3610.34kB 41.02% 41.02%  4163.37kB 47.30%  compress/flate.NewWriter (inline)
  553.02kB  6.28% 47.30%   553.02kB  6.28%  compress/flate.(*compressor).initDeflate (inline)
  532.26kB  6.05% 53.35%   532.26kB  6.05%  github.com/jackc/pgx/v5/pgtype.(*Map).RegisterDefaultPgType (inline)
  516.01kB  5.86% 59.21%   516.01kB  5.86%  hash/crc32.slicingMakeTable
    -514kB  5.84% 53.37%     -514kB  5.84%  bufio.NewReaderSize (inline)
  513.12kB  5.83% 59.20%   513.12kB  5.83%  sync.(*Pool).pinSlow
  512.44kB  5.82% 65.02%   512.44kB  5.82%  vendor/golang.org/x/net/http/httpguts.map.init.0
 -512.23kB  5.82% 59.20%  -512.23kB  5.82%  runtime.mallocgc
 -512.16kB  5.82% 53.38%  -512.16kB  5.82%  encoding/json.NewDecoder (inline)
  512.07kB  5.82% 59.20%   512.07kB  5.82%  net/url.parse
  512.02kB  5.82% 65.02%   512.02kB  5.82%  syscall.(*RawSockaddrAny).Sockaddr
         0     0% 65.02%   513.12kB  5.83%  bufio.(*Writer).Flush
         0     0% 65.02%     -514kB  5.84%  bufio.NewReader (inline)
         0     0% 65.02%   553.02kB  6.28%  compress/flate.(*compressor).init
         0     0% 65.02%  4163.37kB 47.30%  compress/gzip.(*Writer).Write
         0     0% 65.02%   532.26kB  6.05%  database/sql.(*DB).PingContext
         0     0% 65.02%   532.26kB  6.05%  database/sql.(*DB).PingContext.func1
         0     0% 65.02%   532.26kB  6.05%  database/sql.(*DB).conn
         0     0% 65.02%   532.26kB  6.05%  database/sql.(*DB).retry
         0     0% 65.02%  4163.37kB 47.30%  encoding/json.(*Encoder).Encode
         0     0% 65.02%  3651.21kB 41.48%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 65.02%  3651.21kB 41.48%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 65.02%  3651.21kB 41.48%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 65.02%  3651.21kB 41.48%  github.com/go-chi/chi/v5/middleware.StripSlashes.func1
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database.CasRestoreOnErr (inline)
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database.GenerateAdvisoryLockId
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database/postgres.(*Postgres).Lock
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database/postgres.(*Postgres).Lock.func1 (inline)
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database/postgres.(*Postgres).ensureVersionTable
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database/postgres.WithConnection
         0     0% 65.02%   516.01kB  5.86%  github.com/golang-migrate/migrate/v4/database/postgres.WithInstance
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5.connect
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5/pgtype.NewMap
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5/pgtype.initDefaultMap
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5/pgtype.registerDefaultPgTypeVariants[go.shape.[]github.com/jackc/pgx/v5/pgtype.Range[github.com/jackc/pgx/v5/pgtype.Float8]]
         0     0% 65.02%   532.26kB  6.05%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 65.02%   512.02kB  5.82%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 65.02%   516.01kB  5.86%  hash/crc32.ChecksumIEEE
         0     0% 65.02%   516.01kB  5.86%  hash/crc32.archInitIEEE (inline)
         0     0% 65.02%   516.01kB  5.86%  hash/crc32.init.OnceFunc.func4
         0     0% 65.02%   516.01kB  5.86%  hash/crc32.init.OnceFunc.func4.1
         0     0% 65.02%   516.01kB  5.86%  hash/crc32.init.func2
         0     0% 65.02%   513.12kB  5.83%  io.Copy (inline)
         0     0% 65.02%   513.12kB  5.83%  io.CopyN
         0     0% 65.02%   513.12kB  5.83%  io.copyBuffer
         0     0% 65.02%   513.12kB  5.83%  io.discard.ReadFrom
         0     0% 65.02%  1048.27kB 11.91%  main.main
         0     0% 65.02%   512.02kB  5.82%  metrics-collector/internal/app.(*App).Run.func1
         0     0% 65.02%  1048.27kB 11.91%  metrics-collector/internal/app.New
         0     0% 65.02%  3651.21kB 41.48%  metrics-collector/internal/app.registerRoutes.WithLogging.func4.1
         0     0% 65.02%  3651.21kB 41.48%  metrics-collector/internal/app.registerRoutes.WithSignature.func5.1
         0     0% 65.02%  3651.21kB 41.48%  metrics-collector/internal/app.registerRoutes.WithTimeout.func3.1
         0     0% 65.02%  3651.21kB 41.48%  metrics-collector/internal/handler.(*MetricsHandler).UpdateMetricsBatch
         0     0% 65.02%  4163.37kB 47.30%  metrics-collector/internal/middleware.(*gzipResponseWriter).Write
         0     0% 65.02%  3651.21kB 41.48%  metrics-collector/internal/middleware.WithCompressing.func1
         0     0% 65.02%  1048.27kB 11.91%  metrics-collector/internal/repository.NewRepository
         0     0% 65.02%  1048.27kB 11.91%  metrics-collector/internal/repository/postgres.NewPostgresStorage
         0     0% 65.02%   516.01kB  5.86%  metrics-collector/internal/repository/postgres.runMigrations
         0     0% 65.02%   512.02kB  5.82%  net.(*TCPListener).Accept
         0     0% 65.02%   512.02kB  5.82%  net.(*TCPListener).accept
         0     0% 65.02%   512.02kB  5.82%  net.(*netFD).accept
         0     0% 65.02%   512.02kB  5.82%  net/http.(*Server).ListenAndServe
         0     0% 65.02%   512.02kB  5.82%  net/http.(*Server).Serve
         0     0% 65.02%   513.12kB  5.83%  net/http.(*chunkWriter).Write
         0     0% 65.02%   513.12kB  5.83%  net/http.(*chunkWriter).writeHeader
         0     0% 65.02%   512.07kB  5.82%  net/http.(*conn).readRequest
         0     0% 65.02%  4162.40kB 47.29%  net/http.(*conn).serve
         0     0% 65.02%   513.12kB  5.83%  net/http.(*response).finishRequest
         0     0% 65.02%  3651.21kB 41.48%  net/http.HandlerFunc.ServeHTTP
         0     0% 65.02%     -514kB  5.84%  net/http.newBufioReader
         0     0% 65.02%   512.07kB  5.82%  net/http.readRequest
         0     0% 65.02%  3651.21kB 41.48%  net/http.serverHandler.ServeHTTP
         0     0% 65.02%   512.07kB  5.82%  net/url.ParseRequestURI
         0     0% 65.02%   512.44kB  5.82%  runtime.doInit (inline)
         0     0% 65.02%   512.44kB  5.82%  runtime.doInit1
         0     0% 65.02%  1560.71kB 17.73%  runtime.main
         0     0% 65.02%  -512.23kB  5.82%  runtime.malg
         0     0% 65.02%  -512.23kB  5.82%  runtime.newobject
         0     0% 65.02%  -512.23kB  5.82%  runtime.newproc.func1
         0     0% 65.02%  -512.23kB  5.82%  runtime.newproc1
         0     0% 65.02%  -512.23kB  5.82%  runtime.systemstack
         0     0% 65.02%  1048.27kB 11.91%  sync.(*Once).Do (partial-inline)
         0     0% 65.02%  1048.27kB 11.91%  sync.(*Once).doSlow
         0     0% 65.02%   513.12kB  5.83%  sync.(*Pool).Get
         0     0% 65.02%   513.12kB  5.83%  sync.(*Pool).pin
         0     0% 65.02%   512.44kB  5.82%  vendor/golang.org/x/net/http/httpguts.init

## Инкремент 18
Код отформатирован с помощью `gofmt` и `goimports`.