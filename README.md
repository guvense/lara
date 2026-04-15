
# Lara


# Table of Content
- [Overview](#overview)
- [Features](#features)
    * [Handler](#handler)
    * [Callback](#callback)
    * [Dynamic Value](#dynamic-value)
    * [Authorization Interceptor](#authorization-interceptor)
    * [Watcher](#watcher)
 - [Usage](#usage)
 - [Web UI](#web-ui)

## Overview
Lara is a configurable mock server. You just define your specifications as JSON files and let Lara create a server for you.   


# Features
You can check [examples](example) for usage of features.

## Handler
 ### Rest Handler
Rest Handler handles your rest request and return response as you defined.       
- Regex Match     
Lara supports regex for matching requests with mocks thanks to Gorilla mux library. If the regex matched path variable or request parameters, Lara will match request to defined mock specification and return defined response to matched request.      
- Response Delay     
You can also set delay time to simulate real-life scenarios. You only add after field to response or callback objects.     

## Callback
  ### Rest Callback
- You can define one or multiple callback methods for each request. First, Lara will return response according to the spec then Lara invokes your callback action. As mentioned, you can define a delay time for callbacks as well.    

  ### Message Broker Callback
- You can define broker callbacks with `message-broker` for asynchronous publish actions after a response is returned.
- `type` defines broker implementation. Current supported values are `kafka` and `rabbitmq`.
- `after` defines delay before publish operation (example: `10000ms`).
- You can still use dynamic placeholders in broker fields and message payloads.

Example callback definition:
```json
{
  "message-broker": {
    "type": "kafka",
    "after": "10000ms",
    "connection": {
      "brokers": ["localhost:9092"]
    },
    "headers": {
      "x-source": "lara"
    },
    "message": {
      "orderId": "${value::response::body::uuid}"
    },
    "kafka": {
      "topic": "orders.created",
      "key": "${value::request::body::username}"
    }
  }
}
```

RabbitMQ example:
```json
{
  "message-broker": {
    "type": "rabbitmq",
    "after": "5000ms",
    "connection": {
      "url": "amqp://guest:guest@localhost:5672/"
    },
    "message": {
      "event": "user.updated"
    },
    "rabbitmq": {
      "exchange": "events",
      "routing-key": "user.updated"
    }
  }
}
```


## Dynamic Value
Thanks to the dynamic value feature, you can generate unique values for each request. Therefore you can get generic response or send a generic requests to callback services. You only put these scripts to your mock specification.         

### String Scripts
- Random String    
It will generate ten length random string. You can adjust the length of the string by changing the last parameter.     
`${str::random::10} `
- Uuid String    
Generating uuid   
`${str::uuid} `
- Random String From Regex   
You can also generate random strings by defining your own regex format. First, you need you to define your regular expression with using config.yml file as below. You are not allowed to define this regular expression to mock specifications because of parsing rules.       
Otherwise, your specification will behave unexpectedly.
``` config.yml
regex:
  email: "^[a-z]{5,10}@[a-z]{5,10}\\.(com)$"
  number: "/\d+"
``` 
Then you can use this expression with regex keyword. 10 is the length of the string. If you provide the length of the string in your regular expression you don't need to pass string length as parameter. Check the example for more usages.         
`${str::regex::email::10} `

### Number Library
- Random Number      
Generating a random number between -10 and 1    
`${number::generate::-10::1}`


### Date Library
- Date from format   
You can generate date from format    
First you need to define your date format to config file as below.
If you are not provided any format the default format will be;    
`01-02-2006 -> MM-DD-YYYY`


``` 
date:
  simpleDateFormat: "2006-01-02"
  myDateFormat : "2006-02-01"
``` 
Then you can generate dates with the scripts      
`${date::now::simpleDateFormat}`    
`${date::random::simpleDateFormat}`  
The last parameter is the keyword that you added to config file. You can define multiple different date formats to config file with different keywords. You are not allowed to define the format in the mock specification. Check examples for more info.

!!! You need to use go time format time layout. You can check for more info
[Date format](https://gosamples.dev/date-format-yyyy-mm-dd/#:~:text=%F0%9F%93%85%20YYYY-MM-DD%20date%20format%20in%20Go&text=To%2)


### Value keeper 
Lara allowed to you extract values from request and response objects by using value keeper.
For instance, this script will retrieve userId value from the request query parameter.       
`${value::request::queryparams::userId}`         
 You can extract values from     
- Request body    
- Request path variable    
- Request query parameter      
- Response body (you may need this for callbacks)     
For more you can check examples.    

## Authorization Interceptor
If you define callback method and this callback  will send request to a server. In some cases, that server is protected with an authorization method. You can define your authorization credential to config yaml file and use it as value of token-generator into callback specification. Before sending request to the callback server, Lara will first fetch token and use this token while sending request to target service.            
First, you need to define your authorization service credential as below.        
As an example the name of the authorization server is keycloak. You can give any name that you want instead of keycloak.      

config.yml
``` 
token-generator:
   keycloak:
     type: 
     token-url:  
     client-id: 
     client-secret:
     username: 
     password: 
     scope: 
```

moc.json
``` 
  "callback": [
            {
                "rest" : {
                    "after": "3000ms",
                    "token-generator": "keycloak",
                    "request": {
                        "endpoint" : "https://localhost:8443/users/${value::response::body::uuid}/${value::request::queryparams::test}",
                        "method": "POST",
                        "headers" :  {"content-type": "application/json"},
                        "body" : {"username" : "${value::request::body::username}"}
                        }
                }
            },
```

Lara supports Password and Client Credentials grant types. Check example config yml for more.

``` 
type : Password-Credential -> Password Credentials Grant Type
type :  Client-Credential  -> Client Credentials Grant Type
``` 

## Watcher 
If you want to update server when you update mock file, you can basically pass watcher flag to application. Lara watches your
mocks directory and configuration path. If any change happens, Lara will update your server.
There two way to initialize lara with watcher   
First pass wather flag to application   
`--watcher`   
Second add watcher field to config file   
`watcher: true`      

## Web UI
Lara now ships with a separate Next.js web app under `web/`.

- Web app serves editor UI for `config.yml` and `mocks.json`
- Wizard-style flow is supported from UI:
  - `Create Mock Service`
  - Select request type/method/endpoint
  - Configure response details
  - Optionally configure REST callback
  - Optionally configure callback authentication
  - Optionally configure message-broker callback (`kafka` / `rabbitmq`)
  - Apply form output into `mocks.json`, then save
- Backend exposes editor APIs:
  - `GET /api/editor`
  - `POST /api/editor/save`
- Save action writes files and watcher rebuilds Lara server automatically

### Run with Docker Compose

Use one command to start both backend and web UI:

```bash
docker compose up --build
```

If backend image build fails due to `go mod tidy` compatibility resolution, this repository's Dockerfile already uses `go mod download` to avoid mutating module files during image build.

Then open:

- Web UI: `http://localhost:3000`
- Lara backend: `http://localhost:8898`

Compose setup details:

- `backend` service runs Lara with:
  - `--config /data/config.yml`
  - `--mocks /data/mocks`
  - `--watcher=true`
- `web` service proxies editor requests to backend via `LARA_BACKEND_URL=http://backend:8898`
- Shared runtime files are mounted from `./runtime` to backend `/data`

Stop services:

```bash
docker compose down
```

### Docker
- Passing arguments 

```
docker run -p 8899:8899 -v "$PWD/mocks.json:/mocks.json" -v "$PWD/config.yml:/config.yml"  guvense/lara  --config /config.yml --mocks /mocks.json
```


## License
MIT License, see [LICENSE](https://github.com/guvense/lara/blob/main/LICENSE)


Inspired by  [killgrave](https://github.com/friendsofgo/killgrave)
