# notification-service


Ocena 2.0:
- Brak przesłanego rozwiązania.
- Brak zrozumienia stworzonego kodu.
Ocena 3.0:
- System umożliwia tworzenie powiadomień z określną treścią, kanałem (push, e-mail), strefą
czasową i odbiorcą.
- System pozwala użytkownikom planować symulację wysyłki powiadomień na określony
moment w przyszłości, umożliwiając ustawienie dokładnej daty i godziny symulowanego
dostarczenia wiadomości.
- Kanały wysyłki wiadomości działają niezależnie. Obsługa wysyłki kanałów musi odbywać się
na osobnej instancji serwera.
- System gwarantuje, że każda wiadomość zostanie dostarczona dokładnie jeden raz do danego
użytkownika (brak duplikatów i brak pominięć). W przypadku problemów odejmie trzy próby
dostarczenia wiadomości.
- Utworzone powiadomienia są przechowywane przez system do momentu wysłania,
niezależnie od bieżącego obciążenia usługi.
- Proces wysyłki powiadomień odbywa się niezależnie od momentu ich utworzenia, w sposób
uporządkowany i niezawodny.

Uwaga:
- Każdy serwer obsługujący kanał może obsługiwać w tym samym czasie tylko jedną
wiadomość.
- Wysyła wiadomości polega tylko na wyświetleniu w logu odpowiedniej wiadomości. Nie ma
potrzeby implementacji rzeczywistej wysyłki. Każda próba wysyłki ma 50% na to, że się uda.

+ 0.25
System zapewnia punkty końcowe z informacjami o bieżących statusach poszczególnych
powiadomień z wyraźnym oznaczeniem:
- Oczekujące - oczekujące na wysyłkę.
- Wysłane - skutecznie dostarczone.
- Nieudane - niepowodzenie po wszystkich próbach.

+0.5
System uwzględnia lokalne strefy czasowe użytkowników i unika symulowania wysyłania
powiadomień w nieodpowiednich porach (np. symulacja dostarczenia wiadomości poza
nocnymi godzinami lokalnymi użytkownika).

+ 0.25
Administratorzy mają dostęp do statystyk poszczególnych sewerwerów obsługujących kanały
wyświetlanych w punkcie końcowym /metrics,
zawierających liczbę powiadomień wysłanych (symulacja zakończona poprawnie),
oczekujących oraz nieudanych, w zadanym okresie.
Statystki powinny być dostępne osobno dla każdego serwera.

+ 0.25
Przygotowanie konfiguracji kontenerów Docker (Dockerfile oraz docker-compose), która pozwoli
na uruchomienie aplikacji wraz z wszystkimi niezbędnymi komponentami (np. baza danych,
symulowana kolejka wiadomości).

+ 0.25
System umożliwia ustawienie priorytetów powiadomień (nisko / wysoki). Powiadomienia z
wysokim priorytetem powinny mieć większą szanse dotrzeć pierwsze.

+ 0.5
Dodać mechanizm wymuszenia przesłania natychmiast oraz anulowania wysyłki wiadomości
która została zaplanowana.

- 2.0
Kod źródłowy nie spełnia zasad SOLID, DRY oraz KISS lub jest niezgodny z zasadami podziału na
warstwy aplikacji.
