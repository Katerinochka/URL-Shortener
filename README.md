# URL-Shortener

Из соображений экономии времени, предположим, что в день на сервис будет поступать 1 миллион запросов на генерацию короткой ссылки.
Срок годности каждой ссылки будет составлять 1 сутки (дальнейшее развитие проекта предполагает, что за более длительное время будет списываться ещё неустановленная сумма)).
Алфавит состоит из символов латинчкого алфавита в верхнем и нижнем регистре, цифр и символа '_' - всего 63 символа, каждый весит 1 байт.
Длина короткой ссылки (ключа) = 10 символов.
1 ключ = 10 * 1 байт = 10 байт
1 млн коротких ссылок = 1млн * 10байт ≈ 100МБ
Их мы заарнее сгенерируем и будем хранить в сущности свободных ключей.

Когда будет приходить POST-запрос от клиента, заберём один ключ мз пула свободный, отдадим его клиенту и запишем в пул занятых, вместе с оригинальной ссылкой и временем создания этой записи (если при этом вылезет ошибка, скажем об этом клиенту и попросим сделать запрос снова).

Допустим, средняя длина оригинальной ссылки = 500 символов.
Тогда пул занятых ключей = 100МБ (пул коротких) + 1млн * 500 * 1байт (пул оригинальных) + 1млн * 8байт (timestamp) ≈ 600МБ

Когда будет приходить GET-запрос от клиента, найдём в пуле занятых оригинальную ссылку и отправим клиенту (если не найдём, то сообщим об этом клиенту).

Как пользоваться данной API:
#1 POST-запрос:
  Сервер принимает json вида:
    {"longURL": "https://example.com/abra/kadabra"}
  на PATH: "/create"
  Сервер отправляет json вида:
    {"statusmessage":"Ok", "shortURL":"http://ABCD_1234z"}
    
#2 GET-запрос:
  Сервер принимает короткую ссылку в виде параметра:
    "/getoriginal?shorturl=http://ABCD_1234z"
  Сервер отправляет json вида:
    {"statusmessage":"Ok", "shortURL":"https://example.com/abra/kadabra"}
    
Алгоритм генерации:
- для одной ссылки рандомно выбираются и склеиваются 10 символов из заданного алфавита
- каждый раз, когда из пула свободных забирается один ключ, генерируется новый ключ и вставляется в пул
- перед вставкой в пул занятых, проверяется наличие этого ключа в этом пуле. Если такой находится, пользователю отдаётся внутренняя ошибка с просьбой повторить запрос. Если нет, то он попадает в пул занятых.

Очистка просроченных ключей:
Костыль(:
  в отдельном потоке опрашивается текущее время. Когда наступает 00:00:00, запускается проход по пулу занятых, из которого удаляются просроченные ссылки, путём проверки разницы текущего времени и времени создания. Затем поток засыпает на 23 часа. Всё это в бесконечном цикле.
  
Мои команды для запуска контейнера:
1. docker-compose build
2. docker-compose up

Генерация ключей происходит не моментально, окончание генерации подсказывает терминал.

Режим хранилища по умолчанию in-memory. Меняется добавлением флага к запуску -storage db
