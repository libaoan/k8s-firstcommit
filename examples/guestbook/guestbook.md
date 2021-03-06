## GuestBook example

This example shows how to build a simple multi-tier web application using Kubernetes and Docker.

The example combines a web frontend, a redis master for storage and a replicated set of redis slaves.

### Step Zero: Prerequisites
This example assumes that you have forked the repository and turned up a Kubernetes cluster.


### Step One: Turn up the redis master.

Create a file named redis-master.json, this file is describes a single task, which runs a redis key-value server in a container.

```javascript
{
  "id": "redis-master-2",
  "desiredState": {
    "manifest": {
      "containers": [{
        "name": "master",
        "image": "dockerfile/redis",
        "ports": [{
          "containerPort": 6379,
          "hostPort": 6379 
        }]
      }]
    }
  },
  "labels": {
    "name": "redis-master"
  }
}
```

Once you have that task file, you can create the redis task in your Kubernetes cluster using the cloudcfg cli:

```shell
./src/scripts/cloudcfg.sh -c redis-master.json create /tasks
```

Once that's up you can list the tasks in the cluster, to verify that the master is running:

```shell
./src/scripts/cloudcfg.sh list /tasks
```

You should see a single redis master task.  It will also display the machine that the task is running on.  If you ssh to that machine, you can run
```shell
sudo docker ps
```

And see the actual task.  (note that initial ```docker pull``` may take a few minutes, depending on network conditions.

### Step Two: Turn up the master service.
A Kubernetes 'service' is named load balancer that proxies traffic to one or more containers.  The services in a Kubernetes cluster are discoverable inside other containers via environment variables.  Services find the containers to load balance based on task labels.  The task that you created in Step One has the label "name=redis-master", so the corresponding service is defined by that label.  Create a file named redis-master-service.json that contains:

```javascript
{
  "id": "redismaster",
  "port": 10000,
  "labels": {
    "name": "redis-master"
  }
}
```

Once you have that service description, you can create the service with the cloudcfg cli:

```shell
./src/scripts/cloudcfg.sh -c redis-master-service create /services
```

Once created, the service proxy on each minion is configured to set up a proxy on the specified port (in this case port 10000).

### Step Three: Turn up the replicated slave service.
Although the redis master is a single task, the redis read slaves are a 'replicated' task, in Kubernetes, a replication controller is responsible for managing multiple instances of a replicated task.  Create a file named redis-slave-controller.json that contains:

```javascript
  {
    "id": "redisSlaveController",
    "desiredState": {
      "replicas": 2,
      "replicasInSet": {"name": "redis-slave"},
      "taskTemplate": {
        "desiredState": {
           "manifest": {
             "containers": [{
               "image": "brendanburns/redis-slave",
               "ports": [{"containerPort": 6379, "hostPort": 6380}]
             }]
           }
         },
         "labels": {"name": "redis-slave"}
        }},
    "labels": {"name": "redis-slave"}
  }
```

Then you can create the service by running:

```shell
./src/scripts/cloudcfg.sh -c redis-slave-controller.json create /replicationControllers
```

The redis slave configures itself by looking for the Kubernetes service environment variables in the container environment.  In particular, the redis slave is started with the following command:

```shell
redis-server --slaveof $SERVICE_HOST $REDISMASTER_SERVICE_PORT
```

Once that's up you can list the tasks in the cluster, to verify that the master and slaves are running:

```shell
./src/scripts/cloudcfg.sh list /tasks
```

You should see a single redis master task, and two redis slave tasks.

### Step Four: Create the redis slave service.

Just like the master, we want to have a service to proxy connections to the read slaves.  In this case, in addition to discovery, the slave service provides transparent load balancing to clients.  As before, create a service specification:

```javascript
{
  "id": "redisslave",
  "port": 10001,
  "labels": {
    "name": "redis-slave"
  }
}
```

This time the label query for the service is 'name=redis-slave'

Now that you have created the service specification, create it in your cluster with the cloudcfg cli:

```shell
./src/scripts/cloudcfg.sh -c redis-slave-service.json create /services
```

### Step Five: Create the frontend service.

This is a simple PHP server that is configured to talk to both the slave and master services depdending on if the request is a read or a write.  It exposes a simple AJAX interface, and serves an angular based U/X.  Like the redis read slaves it is a replicated service instantiated by a replication controller.  Create a file named frontend-controller.json:

```javascript
  {
    "id": "frontendController",
    "desiredState": {
      "replicas": 3,
      "replicasInSet": {"name": "frontend"},
      "taskTemplate": {
        "desiredState": {
           "manifest": {
             "containers": [{
               "image": "brendanburns/php-redis",
               "ports": [{"containerPort": 80, "hostPort": 8080}]
             }]
           }
         },
         "labels": {"name": "frontend"}
        }},
    "labels": {"name": "frontend"}
  }
```

With this file, you can turn up your frontend with:

```shell
./src/scripts/cloudcfg.sh -c frontend-controller.json create /replicationControllers
```

Once that's up you can list the tasks in the cluster, to verify that the master, slaves and frontends are running:

```shell
./src/scripts/cloudcfg.sh list /tasks
```

You should see a single redis master task, two redis slave and three frontend tasks.

The code for the PHP service looks like this:
```php
<?

set_include_path('.:/usr/share/php:/usr/share/pear:/vendor/predis');

error_reporting(E_ALL);
ini_set('display_errors', 1);

require 'predis/autoload.php';

if (isset($_GET['cmd']) === true) {
  header('Content-Type: application/json');
  if ($_GET['cmd'] == 'set') {
    $client = new Predis\Client([
      'scheme' => 'tcp',
      'host'   => getenv('SERVICE_HOST'),
      'port'   => getenv('REDISMASTER_SERVICE_PORT'),
    ]);
    $client->set($_GET['key'], $_GET['value']);
    print('{"message": "Updated"}');
  } else {
    $read_port = getenv('REDISMASTER_SERVICE_PORT');

    if (isset($_ENV['REDISSLAVE_SERVICE_PORT'])) {
      $read_port = getenv('REDISSLAVE_SERVICE_PORT');
    }
    $client = new Predis\Client([
      'scheme' => 'tcp',
      'host'   => getenv('SERVICE_HOST'),
      'port'   => $read_port,
    ]);

    $value = $client->get($_GET['key']);
    print('{"data": "' . $value . '"}');
  }
} else {
  phpinfo();
} ?>
```

To play with the service itself, find the name of a frontend, grab the external IP of that host from the Google Cloud Console, and visit http://&lt;host-ip&gt;:8080, note you may need to open the firewall for port 8080 using the console or the gcloud tool.