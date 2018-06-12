package main

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3"
)

func main() {

	fmt.Println("=====start=====")

	config := clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	}
	cli, err := clientv3.New(config)
	if err != nil {

		panic(err.Error())
	}

	etcdTest.client = cli
	go etcdTest.watch()

	etcdTest.put()

	etcdTest.lease()

	go etcdTest.keepalive()

	etcdTest.get()

	fmt.Println("======end======")

	select {}
}

var etcdTest EtcdTest

type EtcdTest struct {
	client  *clientv3.Client
	leaseId clientv3.LeaseID
}

func (et *EtcdTest) get() {

	fmt.Println("======get======")

	resp, err := et.client.Get(context.TODO(), "config", clientv3.WithPrefix())
	if err != nil {

		panic(err.Error())
	}

	fmt.Println("get.resphead=>", resp.Header.String())
	fmt.Println("get.count=>", resp.Count)

	for _, msg := range resp.Kvs {

		fmt.Println("get.kv:", string(msg.Key), "=>", string(msg.Value))
	}

	fmt.Println("=====get=end=====")
}

func (et *EtcdTest) put() {

	_, err := et.client.Put(context.TODO(), "sample_key", "sample_value")
	if err != nil {

		panic(err.Error())
	}
}

func (et *EtcdTest) watch() {

	watcher := clientv3.NewWatcher(et.client)
	watchChan := watcher.Watch(context.TODO(), "config", clientv3.WithPrefix())

	for msg := range watchChan {

		fmt.Println("=====watch=====")
		fmt.Println("msg.header=>", msg.Header.String())

		for _, event := range msg.Events {

			switch event.Type {

			case clientv3.EventTypePut:

				fmt.Println("put.kv:", string(event.Kv.Key), "=>", string(event.Kv.Value))
			case clientv3.EventTypeDelete:

				fmt.Println("del.kv:", string(event.Kv.Key), "=>", string(event.Kv.Value))
			default:

				fmt.Println("watch type =>", event.Type)
			}
		}

		fmt.Println("=====watch=end=====")
	}
}

func (et *EtcdTest) lease() {

	lease := clientv3.NewLease(et.client)
	resp, err := lease.Grant(context.TODO(), 5)
	if err != nil {

		panic(err.Error())
	}

	et.leaseId = resp.ID

	_, err = et.client.Put(context.TODO(), "config_ttl_5", "ttl_5", clientv3.WithLease(et.leaseId))
	if err != nil {

		panic(err.Error())
	}

	resp, err = lease.Grant(context.TODO(), 20)
	if err != nil {

		panic(err.Error())
	}

	_, err = et.client.Put(context.TODO(), "config_ttl_20", "ttl_20", clientv3.WithLease(resp.ID))
	if err != nil {

		panic(err.Error())
	}

	onceAliveChan, err := et.client.KeepAliveOnce(context.TODO(), resp.ID)
	if err != nil {

		panic(err.Error())
	}

	fmt.Println("config_ttl_20_onceresp", onceAliveChan.String())
}

func (et *EtcdTest) keepalive() {

	lease := clientv3.NewLease(et.client)
	aliveChan, err := lease.KeepAlive(context.TODO(), et.leaseId)
	if err != nil {

		panic(err.Error())
	}

	for msg := range aliveChan {

		fmt.Println("=====alive=====")
		fmt.Println(msg.String())
		fmt.Println("=====alive=end=====")
	}

}
