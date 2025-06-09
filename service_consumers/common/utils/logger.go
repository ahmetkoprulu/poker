package utils

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *Loggger

type Loggger struct {
	*zap.Logger
	esClient  *elasticsearch.Client
	indexName string
}

func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	zapLogger, err := config.Build()
	Logger = &Loggger{zapLogger, nil, ""}
	if err != nil {
		panic(err)
	}
}

func InitElasticLogger(elasticUrl, serviceName string) {
	url, err := url.Parse(elasticUrl)
	if err != nil {
		panic(err)
	}

	var indexName = url.Query().Get("index")
	password, _ := url.User.Password()
	username := url.User.Username()
	esCfg := elasticsearch.Config{
		Addresses: []string{url.Scheme + "://" + url.Host},
		Username:  username,
		Password:  password,
	}

	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		panic(err)
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(config.EncoderConfig)

	esWriter := &ElasticWriter{client: esClient, indexName: indexName}
	consoleCore := zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), config.Level)
	elasticCore := zapcore.NewCore(encoder, zapcore.AddSync(esWriter), config.Level)

	core := zapcore.NewTee(consoleCore, elasticCore)
	zapLogger := zap.New(core)
	zapLogger = zapLogger.With(zap.String("service", serviceName), zap.String("environment", "test"))
	Logger = &Loggger{Logger: zapLogger, esClient: esClient, indexName: indexName}
}

func (l *Loggger) String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// ElasticWriter implements zapcore.WriteSyncer interface
type ElasticWriter struct {
	client    *elasticsearch.Client
	indexName string
}

func (ew *ElasticWriter) Write(p []byte) (n int, err error) {
	res, err := ew.client.Index(
		ew.indexName,
		strings.NewReader(string(p)),
		ew.client.Index.WithContext(context.Background()),
		ew.client.Index.WithRefresh("true"),
		ew.client.Index.WithDocumentID(strconv.Itoa(int(time.Now().UnixNano()))),
	)

	if err != nil {
		return 0, err
	} else {
		fmt.Println(res)
	}

	return len(p), nil
}

func (ew *ElasticWriter) Sync() error {
	return nil
}
