# CSV Streaming

Berikut adalah tutorial bagaimana cara melakukan processing data tanpa membebani proses dari kinerja api dan jika anda sedang mencari tutorial terkait kafka bisa cek [disini](https://github.com/restuwahyu13/node-kafka), jika anda mebutuhkan sesuatu yang lebih dari yang saya berikan, anda bisa mengcustomisasi nya sesuai dengan kebutuhan anda, untuk data dummy csv filenya anda bisa menggunakannya [disini](https://www.rndgen.com/data-generator).


## User Table
```sql
CREATE EXTENSION "uuid-ossp";

CREATE TABLE public.user (
  id uuid primary key default uuid_generate_v4(),
  first_name varchar(255) not null,
  last_name varchar(255) not null,
  email varchar(255) unique not null,
  gender varchar (255) null
)

CREATE INDEX user_email ON public.user (email)
```