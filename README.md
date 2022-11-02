# Directories Synchronizer

Код в данном репозитории является реализацией финального проекта по курсу **Rebrain Golang Base** (GO-BASE FINAL).

Реализованная программа представляет собой консольную утилиту, осуществляющую однонаправленную синхронизацию
между некоторыми исходной и целевой (копирующей исходную) директориями, пути к которым передаются ей в качестве
параметров.

## Использование программы

Для использования программы необходима утилита `make` (в проекте есть `Makefile`),
которая обычно есть в большинстве *NIX систем.

Перед запуском программы синхронизации необходимо в своей файловой системе выбрать две разные директории:

1) исходную - программа будет читать её содержимое, но не будет модифицировать его, т.е. данная директория
   будет источником для синхронизации;
2) целевую (копирующую исходную) - программа будет как читать, так и активно модифицировать её
   содержимое, т.е. данная директория будет в результате процесса синхронизации стремиться "повторять" содержимое
   исходной.

### Запуск программы

После клонирования репозитория перейдите внутрь его каталога и из него выполните:

`make run srcdir=/path/to/source/dir copydir=/path/to/mirror/dir`

где в параметрах `srcdir` и `copydir` должны быть указаны пути к соответственно исходной директории и целевой
директории-копии.

Данная команда соберёт проект и запустит процесс однонаправленной синхронизации между этими директориями *в фоновом
режиме*.
После этого любая модификация содержимого (файлов и директорий) внутри исходной директории в конечном итоге будет
отражена и в копирующей директории. При этом соответствующие события синхронизации будут автоматически логироваться
(с уровня *INFO* и выше) в локальный файл `tmp/log.txt` каталога с программой.

### Останов программы

Из каталога с программой выполните либо `make stop` для остановки всех процессов синхронизации, которые были запущены с
помощью данной программы (ведь ничто не мешает вам запускать несколько процессов, например для разных пар директорий),
либо же `make stoplast`, если хотите
остановить лишь один процесс синхронизации, запущенный последним.

### Запуск программы в отладочном режиме

Из каталога с программой выполните:

`make debug srcdir=/path/to/source/dir copydir=/path/to/mirror/dir`

где в параметрах `srcdir` и `copydir` должны быть указаны пути к соответственно исходной директории и целевой
директории-копии.

Данная команда запустит процесс однонаправленной синхронизации между этими директориями с некоторыми включёнными
debug-опциями (в частности, будет включён race-детектор, а логирование будет идти в саму консоль с уровня *DEBUG* и
выше) и уже не в фоновом режиме, а в *foreground* вашей консоли. Для останова вам достаточно будет нажать *Ctrl+C* в
вашей консоли.

### Прогон тестов

Из каталога с программой выполните `make test` для прогона всех тестов в проекте. Эта команда также отобразит
примерное итоговое тестовое покрытие кода.

### Запуск бенчмарка

Из каталога с программой выполните `make bench` для запуска бенчмарка для функции копирования файла.

### Прочие возможности

- Для сборки проекта (без запуска программы) выполните `make build`.
- Для вывода списка всех запущенных процессов синхронизации (запущенных с помощью данной программы)
  выполните `make getpid`.
- Для отображения справки по всем командам (т.е. по всем доступным make targets) выполните `make help` или просто `make`
  .

## Замечания по программной реализации

### Настройки программы

Внутри программы имеется определённый набор настроек, некоторые из которых явно указываются при реализации make targets
(см. в `Makefile`), а некоторые не указываются явно, и задействуются их значения по умолчанию.

Тем не менее, если запускать собранный бинарник с программой (для сборки можно использовать `make build`) напрямую без
`make`, то с помощью флагов и аргументов команды запуска все эти настройки можно явно задавать. Для этого внутри
используется пакет `flag` стандартной библиотеки. В частности, есть и такие потенциально полезные параметры настроек:

- `-hidden` - нужно ли синхронизировать в т.ч. т.н. скрытые файлы и директории (имена которых начинаются с .), по
  умолчанию `false`;
- `-once` - позволяет запустить программу для "разовой" синхронизации, т.е. выполнится один синхро-цикл и программа
  довольно быстро завершится, по умолчанию `false`;
- `-pid` - выводить ли в консоль при запуске программы PID запущенного процесса, по умолчанию `false`;
- `-scanperiod` - период "сканирования" исходной директории в ходе процесса синхронизации, по умолчанию 1 секунда;
- `-workers` - размер пула горутин, выполняющих собственно сами синхронизационные операции, по умолчанию
  равен `runtime.NumCPU()`;
- `-loglvl` - для задания уровня логирования, по умолчанию *INFO*.

### Использованные внешние зависимости

В проекте использовано минимальное число зависимостей, а именно: логгер (**zap**), и библиотеки, используемые для тестов
(**stretchr/testify** и **golang/mock**).

Для удобства все зависимости проекта уже "завендорены" в репозитории.

### Логирование

При логировании есть 4 уровня: *DEBUG, INFO, WARN, ERROR*. При обычном запуске (`make run ...`) активен уровень *INFO*,
а при запуске в режиме отладки (`make debug ...`) - уровень *DEBUG*. Разумеется, при выборе определённого уровня будут
логироваться сообщения и более высоких уровней.

В лог пишутся: уровень (см. выше), время записи, id и тип синхронизационной операции, частичный (относительный) путь к
синхронизируемому объекту (файлу или директории). А также: время, когда операция была запланирована, когда начата и
когда успешно завершена или отменена (например, до её завершения выяснилось, что она уже не нужна).

Логирование осуществляется в локальный файл `tmp/log.txt` (относительно каталога с программой). Этот файл один, его
местоположение не меняется, какая-либо ротация не реализована.

В настройках есть опция, позволяющая писать лог прямо в консоль вместо файла, и эта опция используется при запуске
программы в режиме отладки. 
