# lara



# Table of Content
- [Overview](#overview)
- [Features](#features)
    * [Handler](#handler)
    * [Callback](#callback)
    * [Dynamic Value](#dynamic-value)
    * [Authorization Interceptor](#authorization-interceptor)
    * [Watcher](#watcher)
 - [Usage](#usage)

## Overview
Lara is a mock server. You only define your server rules as json files and let lara create a server for you. 

# Features
You can check [examples](mocs) for usage of features.

## Handler
 ### Rest Handler
- Regex Match     
Lara supports regex for matching request with mocks thanks to gorilla mux library.
- Response Delay     
You can adjust response time with usign after term.

## Callback
  ### Rest Callback
- You can define callback method.First lara will return response according to spec then lara invoke your action.


## Dynamic Value
Thanks to dynamic value, you can generate values for each request. Therefore you can get generic response or send generic request to callback services.

### String Library
- Random String      
`${str::random::10} `
- Uuid String     
`${str::uuid} `

### Number Library
- Random Number      
Generating a random number between -10 and 1    
`${number::generate::-10::1}`


### Value keeper 
You can extract values from request and response with using value keeper.     
For instance this script will retrieve user from request query parameter. For more you can check examples.     
`${value::request::queryparams::userId}`


## Authorization Interceptor
If you define callback method and this callback will send request to a server. A server might wait for authorization token. You can define your authorization credential to config yaml file and use it as value of token-generator. Before sending request to callback server, lara will first fetch token and use this token while fetching callback service.       
First you need to define your authorization service credential as below.      
As an example keycloak will our token server name. You can give any name that you want instead of keycloak.    
type : Password-Credential, Client-Credential      
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
                        "body" : {"username" : "${value::request::body::username}", "pitircik" : "${str::uuid}"}
                        }
                }
            },
```
Lara only support Password Credentials for now.

## Watcher 
If you want to update server when you update mock file, you can basically pass watcher flag to application.    
--watcher flag or config

## Usage 

### Docker
- Passing arguments

``` 
docker run -p 8898:8898 -v "$PWD/mocs:/mocks" guvense/lara  --mocks /mocks --host 0.0.0.0
```

- Using config file
```
sudo docker run -p 8899:8899 -v "$PWD/mocs:/mocks" -v "$PWD/config.yml:/config.yml"  guvense/lara  --config /config.yml
```
