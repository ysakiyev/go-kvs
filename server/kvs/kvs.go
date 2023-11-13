package kvs

type Kvs struct {
	storage map[string]string
}

func New() *Kvs {
	return &Kvs{
		storage: make(map[string]string),
	}
}

func (k *Kvs) Get(key string) string {
	return k.storage[key]
}

func (k *Kvs) Set(key, val string) string {
	k.storage[key] = val
	return k.storage[key]
}

func (k *Kvs) Del(key string) {
	delete(k.storage, key)
}
