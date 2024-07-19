# Set up

```bash
git clone https://github.com/husky-parul/kepler.git
cd kepler
git fetch origin -a
git switch -c local-dev origin/local-dev-docker-compose 
cd compose/validation/metal/
docker compose up
```


### Ensure the setup works broswer:

>Note: This will not work on peaks env because we don't have console. If you running on Mac or Linux machine you use this to verify

- opening an browser
- localhost:9090
- click on Status -> Targets. 
- verify that kepler endpoint http://kepler-metal:8888/metrics is present and the status is UP
- You can check the list of metrics by going to Graph and querying. Example search for kepler_node_platform_joules_total

### Ensure the setup is working via commandline.

Incase you do not open a browser on the server you can also check via command line if kepler endpoint is up

```bash
curl --request GET  localhost:9090/api/v1/query?query=up

{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9090","job":"prometheus"},"value":[1721417625.093,"1"]},{"metric":{"__name__":"up","instance":"kepler-metal:8888","job":"metal"},"value":[1721417625.093,"1"]}]}}[peaks@kube-master-90 ~]$ 


```