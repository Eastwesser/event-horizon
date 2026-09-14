prosto — 26.05.2026 10:31вторник, 26 мая 2026 г. 10:31
Удаление дублей. Реализовать запрос на удаление повторяющихся значений. 
Например есть таблица emails(id, email), в какой-то момент email'ы стали дублироваться, 
необходимо написать запрос на удаление дублей (какие id останутся не принципиально)

create table email(
id bigserial,
email text
);

insert into email(id, email) values (1, 'ivanov@mail.ru');
insert into email(id, email) values (2, 'petrov@mail.ru');
insert into email(id, email) values (3, 'ivanov@mail.ru');
insert into email(id, email) values (4, 'sidorov@mail.ru');
insert into email(id, email) values (5, 'orlov@mail.ru');
insert into email(id, email) values (6, 'sidorov@mail.ru');


DELETE 
FROM email
WHERE id NOT IN 
(SELECT MIN(id)
FROM enail
GROUP BY email);
:thumbsup:
Нажмите, чтобы отреагировать
:orange_heart:
Нажмите, чтобы отреагировать
:grin:
Нажмите, чтобы отреагировать
Добавить реакцию
Ответить
Переслать
Ещё
[10:33]вторник, 26 мая 2026 г. 10:33
Дана строка (возможно, пустая), состоящая из букв A-Z:
AAAABBBCCXYZDDDDEEEFFFAAAAAABBBBBBBBBBBBBBBBBBBBBBBBBBBB
Нужно написать функцию RLE, которая на выходе даст строку вида:
A4B3C2XYZD4E3F3A6B28 "         "
И сгенерирует ошибку, если на вход пришла невалидная строка.
Пояснения:
Если символ встречается 1 раз, он остается без изменений;
Если символ повторяется более 1 раза, к нему добавляется количество повторений.


то что нарешала на собесе, есть ошибки, для них был важен алгоритм
public statis String rle(String input) {
if (input == null || input.isEmpty()) {
throw new IllegalArgumentException("Невалидная строка");
}

StringBuilder result = new StringBuilder();
int cont = 0;
char current = input.charAt(0);

for(int i = 1, i < input.lenght(), i ++) {
char next = input.charAt(i);

if (current = next) {
count ++;
} else {
addElement(result, current, cont)
current = next;
count = 1;
}

addElement(result, current, cont)

return result.toString();
}


private statis addElement(StringBuilder result, char current, int cont) {
result.append(current);
if (count > 1) {
result.append(count);
}
}

} (изменено)вторник, 26 мая 2026 г. 10:33
:thumbsup:
Нажмите, чтобы отреагировать
:orange_heart:
Нажмите, чтобы отреагировать
:grin:
Нажмите, чтобы отреагировать
Добавить реакцию
Ответить
Переслать
Ещё

prosto — 26.05.2026 10:41вторник, 26 мая 2026 г. 10:41
что в итоге выполниться

@Service
class A {
 
    // исходный вызов приходит сюда
    @Transactional
    public void doA() {
        // some work with DB A
 
        try {
            doB();
        } catch (Exception ex) {
            // just log
        }
    }

    @Transactional(propagation = Propagation.NEVER)
    public void doB() {
        // some work with DB B
 
        if (true) { // some operation with exception
            throw new IllegalStateException();
        }
    }
}


class User {
    
    String name;
    String surname;
    
    public User(String a, String b) {
        name = a;
        surname = b;
    }
    
}

class MyCode {
    public static void main(String... args) {
        // Есть список пользователей, Анна, Анна, Мария, Петр, Петр
        // Должны остаться Анна, Мария, Петр
        List<User> users = new ArrayList<User>();
        User user1 = new User("Анна",  "Петрова");
        User user2 = new User("Анна",  "Иванова");
        User user3 = new User("Анна",  "Сидорова");
        User user4 = new User("Мария",  "Петрова");
        User user5 = new User("Петр",  "Сидоров");
        User user6 = new User("Петр",  "Иванов");
        users.add(user1);
        users.add(user2);
        users.add(user3);
        users.add(user4);
        users.add(user5);
        users.add(user6);
       List<User> result = getUniqNameUsers(users);
    }
      
:thumbsup:
Нажмите, чтобы отреагировать
:orange_heart:
Нажмите, чтобы отреагировать
:grin:
Нажмите, чтобы отреагировать
Добавить реакцию
Ответить
Переслать
Ещё
[14:13]вторник, 8 июля 2025 г. 14:13
вывести список не повторяющихся имен
17 февраля 2026 г.

Kostyan — 17.02.2026 16:04вторник, 17 февраля 2026 г. 16:04
 java
# Дан массив чисел в котором все числа кроме одного имеют пару(встречаются дважды)
# Найти число, которое встречается только один раз

# <= [1, 0, 3, -2, 9, 9, 1, -2, 0]
# => 3

# <= [1]
# => 1

# <= [1, 0, 3]
# => 1/0/3 ? ex

# <= [1, 1, 1]
# => 1 ? ex

# нет числа без пары => ex +
# пустой массив + null => ex +
# более 1 числа без пары => ex +
# число повторяется более 2 раз => ex +
int find(int[] arr) {
  if (arr == null || arr.length == 0) {
    throw new IllegalArgumentException("")
  }
    
  Map<Integer, Integer> countMap = new HashMap<>();
  
  for(int num : arr) {
      countMap.put(num, countMap.getOrDefault(num, 0) + 1);
  }
  
  Integer result = null;
  for(Map.Entry<Integer, Integer> entry : countMap.entrySet()) {
      if (entry.getValue() > 2) {
        throw new IllegalArgumentException("") 
      }
    
      if (entry.getValue() == 1) {
          if (result != null) {
              throw new IllegalArgumentException("") 
          }
          result = entry.getKey();
      }
  }
  
  if (result == null) {
      throw new NoSuchElementException("не найден")
  }
  
  return result;
  
}

1
Добавить реакцию
:thumbsup:
Нажмите, чтобы отреагировать
:orange_heart:
Нажмите, чтобы отреагировать
:grin:
Нажмите, чтобы отреагировать
Добавить реакцию
Ответить
Переслать
Ещё
[16:05]вторник, 17 февраля 2026 г. 16:05
с граничными случаями с решением
:thumbsup:
Нажмите, чтобы отреагировать
:orange_heart:
Нажмите, чтобы отреагировать
:grin:
Нажмите, чтобы отреагировать
Добавить реакцию
Ответить
Переслать
Ещё
26 мая 2026 г.
НОВОЕ

DungeonMaster — 26.05.2026 15:01вторник, 26 мая 2026 г. 15:01
1) написать простейший метод, который бы при работе бросил StackOverflowError
2) написать свой класс, который бы реализовывал стек,
с методами push, pop,  и peekMax, который бы возращал max Элемент в стеке за О(1)
3) Есть система, которая дает юзерам возможность работать с файлами в браузере.
Стек стандартный: Java, Spring, React, Postgres. Файлы хранятся в файловой системе на бэке,
метаданные файлов в БД. Команда реализовала фичу - переименование файла
@Transactional
public void process(String oldName, String newName) {
    Long id = exec("select id from file where name='" + oldName + "'"); //выполнение запроса к БД
    processFile(oldName, newName); //переименование файла на диске
    exec("update file set name='" + newName + "' where id = " + id); //выполнение запроса к БД
}

Изображение 1: SQL-запросы
sql
create table users( id int primary key, fio varchar(100) not null, int age)

create index fio on users(fio)

select fio from users group by fio having count(*)>1
Изображение 2: Java-код (Удаление дубликатов)
Условие:

java
//Удаление элементов из списка, начиная с 3 повторения

// Input: ['A', 'B', 'A', 'B', 'A', 'B', 'C', 'C', 'D', 'C', 'C']
// Output: [A, B, A, B, C, C, D]
//List<Character> removeDuplicates(List<Character> elements);
Код:

java
import java.util.*;

public class Main {
    public static void main(String[] args) {
        ArrayList<Character> arr = new ArrayList<>(List.of('A', 'B', 'A', 'B', 'A', 'B', 'C', 'C', 'D', 'C', 'C'));
        removeDuplicates2(arr);
        System.out.println(arr);
    }
    static List<Character> removeDuplicates(List<Character> list) {
        Map<Character,Integer> map = new HashMap<>();

        List<Character> result = new ArrayList<>();

        for(Character ch: list){
            int count = map.getOrDefault(ch, 0);

            if(count<2){
                result.add(ch);
            }

            map.put(ch,++count);
        }
        return result;
    }
    static void removeDuplicates2(ArrayList<Character> list) {
        Map<Character,Integer> map = new HashMap<>();
        Iterator<Character> iterator = list.iterator();
        while(iterator.hasNext()){
            char ch = iterator.next();
            int count = map.getOrDefault(ch, 0);

            if(count<2){
                map.put(ch,++count);
            } else{
                iterator.remove();
            }
        }
    }
}


