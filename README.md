<p align="center">
  <img src="https://cdn.aahframework.org/assets/img/aah-logo-64x64.png" />
  <h2 align="center">Inmemory Cache Provider by aah framework</h2>
</p>

High performance, eviction modes (TTL, NoTTL, Slide), goroutine safe inmemory cache provider. Library keeps cache entries on heap but omits GC and without impact on performance by applying trick of [go1.5 map non-pointer values](https://github.com/golang/go/issues/9477).

### News

  * `v0.1.0` [released](https://github.com/aahframework/inmemory-cache-provider/releases/latest) and tagged on TBD.

## Installation

```bash
go get -u aahframework.org/cache/inmemory
```

Visit official website https://aahframework.org to learn more about `aah` framework.
