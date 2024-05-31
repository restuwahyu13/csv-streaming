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

## Factors Affecting Performance

- Banyaknya jumlah csv data
- Koneksi jaringan internet
- Hardware device
- Database koneksi
- Condingan & SQL Query
- Rebalancing consumer (Kafka)
- Etc

## Tested By Me

Seperti yang saya bilang jika anda upload **1JT** csv data nanti akan dibagi dengan `CHUNK_CSV_SIZE`, contoh jadi **1JT / 100RB** berarti nanti bakal ada **10 Process Batch** yang dimana nanti akan di publish ke `kafka`, anda bisa sesuaikan `CHUNK_CSV_SIZE` dan `GORUTINE_POOL_SIZE` sesuai kebutuhan anda masing - masing, jika dirasa data yang di upload semakin besar anda bisa increase `GORUTINE_POOL_SIZE`, perlu di ingat jangan terlalu besar karena bisa memperlambat dan terlalu banyak menggunakan `cpu`. saya testing `300RB` sampai `500RB` data itu masih terbilang lumayan agak cepat dan low consuming memory walau cpu agak sedikit naik.

## Consumer Group

Anda bisa menggunakan consumer group dengan cara memanggil seperti berikut ini, dan menaruh nya di consumer

```go
	if err := w.broker.ConsumerGroup(packages.TCP, consumerTopicName, consumerGroupName); err != nil {
		return res, err
	}
```