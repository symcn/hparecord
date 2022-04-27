# hparecord

Record hpa's metrics values to prometheus

## Usage

### build

```shell
make docker-build
make docker-psuh
```

### install

```shell
make deploy
```

if you don't have prometheus-operator, just add this job to your prometheus
to scrape metrics
```yaml
      - job_name: hpa
        scrape_interval: 30s
        scrape_timeout: 15s
        static_configs:
          - targets:
            - hparecord:80
```