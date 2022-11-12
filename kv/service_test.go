package kv_test

import (
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/weplanx/utils/kv"
	"sync"
	"testing"
	"time"
)

func TestLoadExistsValues(t *testing.T) {
	err := service.Load()
	assert.NoError(t, err)
	err = service.Load()
	assert.NoError(t, err)
}

func TestLoadBadValues(t *testing.T) {
	_, err := keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)
	err = service.Load()
	assert.Error(t, err)
}

func TestLoadBucketCleared(t *testing.T) {
	err := keyvalue.Delete("values")
	assert.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := service.Load()
		assert.Error(t, err)
		wg.Done()
	}()
	go func() {
		// 误执行
		err := js.DeleteKeyValue("dev")
		assert.NoError(t, err)
	}()
	wg.Wait()
	err = service.Load()
	assert.Error(t, err)
}

func TestSyncBucketCleared(t *testing.T) {
	err := service.Sync(nil)
	assert.Error(t, err)
}

func TestSync(t *testing.T) {
	var err error
	keyvalue, err = js.CreateKeyValue(&nats.KeyValueConfig{Bucket: "dev"})
	assert.NoError(t, err)

	option := kv.SyncOption{
		Updated: make(chan *kv.DynamicValues),
		Err:     make(chan error),
	}
	go func() {
		err := service.Sync(&option)
		assert.NoError(t, err)
	}()
	time.Sleep(time.Millisecond * 500)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		times := 0
		for {
			if times == 2 {
				break
			}
			select {
			case x := <-option.Updated:
				if times == 0 {
					assert.Equal(t, kv.DEFAULT.LoginTTL, x.LoginTTL)
					assert.Equal(t, kv.DEFAULT.LoginFailures, x.LoginFailures)
					assert.Equal(t, kv.DEFAULT.IpLoginFailures, x.IpLoginFailures)
					assert.Equal(t, kv.DEFAULT.PwdStrategy, x.PwdStrategy)
					assert.Equal(t, kv.DEFAULT.PwdTTL, x.PwdTTL)
					assert.Equal(t, "", x.Office)
				}
				if times == 1 {
					assert.Equal(t, kv.DEFAULT.LoginTTL, x.LoginTTL)
					assert.Equal(t, kv.DEFAULT.LoginFailures, x.LoginFailures)
					assert.Equal(t, kv.DEFAULT.IpLoginFailures, x.IpLoginFailures)
					assert.Equal(t, kv.DEFAULT.PwdStrategy, x.PwdStrategy)
					assert.Equal(t, kv.DEFAULT.PwdTTL, x.PwdTTL)
					assert.Equal(t, "feishu", x.Office)
				}
				times++
			case e := <-option.Err:
				assert.Error(t, e)
				times++
			}
		}
		wg.Done()
	}()

	err = service.Set(M{
		"office": "feishu",
	})
	assert.NoError(t, err)

	_, err = keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)

	wg.Wait()
}

func TestSetBadValues(t *testing.T) {
	_, err := keyvalue.Put("values", []byte("abc"))
	assert.NoError(t, err)

	err = service.Set(M{})
	assert.Error(t, err)
}
