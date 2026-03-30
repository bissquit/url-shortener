# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
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

## Профилирование памяти (pprof)

### 1. Написание бенчмарков

Добавлены бенчмарки с `b.ReportAllocs()` для хендлеров:
- `BenchmarkCreateJSON`
- `BenchmarkCreate`
- `BenchmarkRedirect`
- `BenchmarkGetUserURLs`

Добавлены бенчмарки для in-memory хранилища:
- `BenchmarkCreate`
- `BenchmarkGetURLByID`
- `BenchmarkCreateBatch`
- `BenchmarkGetURLsByUserID`

### 2. Снятие базового профиля (до оптимизации)

```bash
go test -bench=. -benchmem -memprofile=profiles/base.pprof ./internal/handler/
```

Результат:
```
BenchmarkCreateJSON-8      576966     2073 ns/op    8166 B/op    41 allocs/op
BenchmarkCreate-8          686746     1668 ns/op    7810 B/op    36 allocs/op
BenchmarkRedirect-8       1350764      885 ns/op    6050 B/op    17 allocs/op
BenchmarkGetUserURLs-8     634707     1897 ns/op    7230 B/op    36 allocs/op
```

### 3. Анализ профиля

```bash
go tool pprof -top profiles/base.pprof
```

Неэффективные места:
- `handler.(*URLHandlers).Create` — тройная конвертация `string(body)`
- `memory.(*URLStorage).GetURLsByUserID` — слайс без pre-allocation растёт через `append`

### 4. Оптимизация кода

Выполнена оптимизация проблемных мест.

### 5. Снятие профиля после оптимизации

```bash
go test -bench=. -benchmem -memprofile=profiles/result.pprof ./internal/handler/
```

Результат:
```
BenchmarkCreateJSON-8      578682     2054 ns/op    8166 B/op    41 allocs/op
BenchmarkCreate-8          740031     1654 ns/op    7714 B/op    34 allocs/op
BenchmarkRedirect-8       1371513      872 ns/op    6050 B/op    17 allocs/op
BenchmarkGetUserURLs-8     628552     1913 ns/op    7230 B/op    36 allocs/op
```

`BenchmarkCreate`: 36 -> 34 allocs/op, 7810 -> 7714 B/op.

### 6. Сравнение профилей

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```
File: handler.test
Type: alloc_space
Showing nodes accounting for 88.19MB, 0.29% of 30462.68MB total
      flat  flat%   sum%        cum   cum%
  -72.50MB  0.24%  0.24%    61.03MB   0.2%  handler.(*URLHandlers).Create
   50.02MB  0.16% 0.074%    50.02MB  0.16%  net/http.(*Request).WithContext
     -49MB  0.16%  0.23%      -49MB  0.16%  net/http/httptest.NewRecorder
   41.02MB  0.13%   0.1%    41.02MB  0.13%  io.ReadAll
   38.52MB  0.13% 0.026%    38.52MB  0.13%  net/http.Header.Clone
   37.14MB  0.12%  0.15%    37.14MB  0.12%  bufio.NewReaderSize
  -37.01MB  0.12% 0.027%   -37.01MB  0.12%  net/textproto.MIMEHeader.Set
   33.50MB  0.11%  0.14%    33.50MB  0.11%  net/url.parse
   31.01MB   0.1%  0.24%    49.01MB  0.16%  net/http.readRequest
  -14.50MB 0.048%  0.19%   -14.50MB 0.048%  encoding/json.NewDecoder
     -11MB 0.036%  0.28%   -37.51MB  0.12%  handler.(*URLHandlers).GetUserURLs
      -2MB 0.0066%  0.28%      -11MB 0.036%  handler.(*URLHandlers).CreateJSON
```

Отрицательные значения (`-72.50MB` в `Create`, `-37.51MB` cum в `GetUserURLs`) подтверждают уменьшение потребления памяти после оптимизации.
