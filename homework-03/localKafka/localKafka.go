package localKafka

import (
	"container/list"
	"context"
	"localRedis"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

const (
	urlTopic       = "aisakova-tinyurls"
	broker1Address = "158.160.19.212:9092"
)

func CreateUrlWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker1Address},
		Topic:        urlTopic,
		RequiredAcks: 1,
	})
}

func CreateUrlReaders(nWorkers int) *list.List {
	urlReaders := list.New()
	for i := 0; i < nWorkers; i++ {
		urlReaders.PushBack(kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker1Address},
			Topic:   urlTopic,
		}))

	}
	return urlReaders

}

func UrlProduce(writer *kafka.Writer, ctx context.Context, longUrl string, tinyUrl string) {
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(tinyUrl),
		Value: []byte(longUrl),
	})
	if err != nil {
		panic("could not write message " + err.Error())
	}
	//fmt.Println("writes:", tinyUrl+":"+longUrl)

}

func UrlConsume(reader *kafka.Reader, ctx context.Context, cluster *localRedis.RedisCluster, id int) {
	(*cluster).RedisOptions.Addr = (*cluster).Workers[id]
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		rdb.Do(ctx, "set", (*cluster).Prefix+"_"+string(msg.Key), string(msg.Value))
		//fmt.Println("received: ", string(msg.Value))

	}
}
