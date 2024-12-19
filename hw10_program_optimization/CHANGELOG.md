# Ход работ

# Начальное состояние
коммит c3b27f5404712e8cc8891517ce7e771d5f72b86f

- cpu.prof.orig
- mem.prof.orig

```bash
   stats_optimization_test.go:46: time used: 577.097228ms / 300ms
   stats_optimization_test.go:47: memory used: 308Mb / 30Mb
```

Работает долго, много памяти.

# Анализ кода
Изучаю код.

```go
content, err := io.ReadAll(r)

...
lines := strings.Split(string(content), "\n")
.....	
	
type users [100_000]User	
```

Проблемы:
- В коде сначала вычитывается ВЕСЬ файла в память, затем он делится по '\n'.
- В коде все данные пользователей записываются в одну структуру (а что если данных будет больше?). 
- Затем идет обработка данной структуры в цикле.

Идея:
- работать с кусками файла до разделителя '\n'.
- за одно решил поэкспериментировать с easyjson.
- напрашивается разделение функционала по горутинами - одна считает данные, другая проверяет на domain, третья считает сумму

Доработки:
- коммит 2693091100d00e8a7adcc94fe9b7a42caf9ceedd (scanner + easy json)
- коммит bbfe22d014010e46ab1c3b99f11358f1ed6cd326 (убрал использование общего списка пользователей, разделил на функции по каналам)

```bash
   stats_optimization_test.go:46: time used: 421.339646ms / 300ms
   stats_optimization_test.go:47: memory used: 146Mb / 30Mb
```

Программа перестала зависеть от размера исходных данных, хотя по перфомансу изменений нет. Изучаю prof. (cpu_prof.sh, mem_prof.sh)

- cpu.prof.goroutines
- mem.prof.goroutines
 
Вижу что много уходит на вызов regex.compile  (cpu.prof.goroutines.svg)

Идея:
- добавить кэш для регулярок

Доработки:
- коммит 2c8dd70530c41b50996735d863d92ee60cc64585 (кеш для регулярок)

```bash
   stats_optimization_test.go:46: time used: 464.357625ms / 300ms
   stats_optimization_test.go:47: memory used: 10Mb / 30Mb
```

- cpu.prof.regcache
- mem.prof.regcache

Удивительно - быстрее не стало, зато использование памяти уменьшилось. 

Изучаю  cpu.prof.regcache ( cpu.prof.regcache.svg).

Вижу что больше всего времения уходит на вызов bufio.Scanner. scanner создается из io.Reader. 

Идея:
- добавить буферизацию ридера.

Доработка:
- коммит 781a4dd4e223e5ff5a38306407f053bb501ab7e2 (обернул reader в bufio)
- 
```bash
   stats_optimization_test.go:46: time used: 209.675046ms / 300ms
   stats_optimization_test.go:47: memory used: 10Mb / 30Mb
```
- cpu.prof.bufio
- mem.prof.bufio

**Тесты пройдены!!**

- по cpu уперлись в работу декомпрессора и jlexer
- по mem уперлись в jlexer


Решил попробовать заменить regexp на strings.HasSuffix

- коммит ab68397a2a415dbaecb07edf00e31a01417559a1

```
    stats_optimization_test.go:46: time used: 239.420839ms / 300ms
    stats_optimization_test.go:47: memory used: 14Mb / 30Mb
```

- cpu.prof.suffix
- mem.prof.suffix

Результат оказался хуже чем в 781a4dd4e223e5ff5a38306407f053bb501ab7e2; ревертнул коммит.




