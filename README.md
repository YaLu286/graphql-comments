# graphql-comments

### Приветствую вас, доблестная команда

#### Небольшое предисловие:

+ В своей реализации сервиса я отказался от использования вложенных массивов в типах данных, решив для себя сразу множество проблем с оптимизацией.
+ Поле Replies в типе Post получает исключительно верхний уровень комментариев.
+ Далее при необходимости мы можем получить нужный тред, спустившися на уровень в иерархии, 
используя запрос Comments с возможность кейсет пагинации, мы можем знать на какой из комментариев были ответы и какой parentID указывать в очередном запросе, так как в типе Comment есть поле hasReplies
+ структура бд и хранения данных памяти ясна из sql скрипта и комментариев в пакете *storage/inmemory*
+ небольшой пакет *config* призван помочь с настройкой нашего сервиса с помощью переменных окружения
+ пакет *graph* содержит имплементацию резольверов и файлы и модели, сгенерированные с помощью gqlgen от 99designs
+ в пакете *storage* описан интерфейс хранилища
+ в пакетах *storage/inmemory* и *storage/postgres* описаны реализации этого интерфейса для хранения в памяти и в бд соответственно
+ в *models* разумеется описаны типы Post и Comment и (де)сериализаторы типов ID и Timestamp
+ в корневой директории проекта есть Dockerfile для сборки образа нашего сервера
+ а также docker-compose.yml, чтобы можно было набрать docker compose up --build и вуаля. На http://localhost:{port} можно потестить всё в красочной песочнице. В http://localhost:{port}/query можно покидать запросы c помощью curl/Postman