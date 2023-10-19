Soil stands for System Orchestration and Investigation Lever, is a tool for
deploying an environment and running tests.

The soil is a wrapper around terraform, which help to manage multiple deployments
and track the states locally.

Deploy the configuration

```shell
so deploy
so remove

so [deploy|remove]


so [deploy|dp] [NAME] [options] create deployment
so [remove|rm] [NAME] [options] remove deployment
so [status|st] [NAME] [options] lookup deployment
so [test|st] [NAME] [otions] run deployment secific tests
```


Deployment types
----------------

[-t|--type] aws|k3d|ssh

System requirements
-------------------

The following tools must be present in the system for `k3d` deployment type:

- kubectl
- helm
- Terraform
- git

By default k3d requires read access to terraform and k6 scalability tests repo: https://github.com/moio/scalability-tests

For custom branch please use repo specific option `-r`:

```shell
so deploy -r https://github.com/moio/scalability-tests@yourbranch
```

Development guide
-----------------

For deploying k3d based environment and run tests it is enough following commands:

```shell
git clone https://github.com/.../soil

cd soil
# deploy default k3d based environment
go run . deploy
# run k9 tests
go run . test
# cleanup
go run . rm -f
```

Building being in `soil` repo root directory:

```shell
go build -o so .
```

Link to home bin:
```shell
(cd $HOME/bin ; ln -s $PWD/so so)
```

