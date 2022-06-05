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

## Overview
Lara is a mock server. You only define your server rules as json files and let lara create a server for you. 

# Features
You can check [examples](example) for usage of features.

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
- Random String From Regex   
You can also generate random string from regex. First you need you define your regular expression to config.yml file as below
``` 
regex:
  email: "^[a-z]{5,10}@[a-z]{5,10}\\.(com|net|org)$"
  number: "/\d+"
``` 
Then you can use this expression with regex keyword. 10 is the lenght of string.
`${str::regex::email::10} `

### Number Library
- Random Number      
Generating a random number between -10 and 1    
`${number::generate::-10::1}`


### Date Library
- Current Date from format   
You can generate current date from format    
First you need to define your date format to config file as below
``` 
date:
  simpleDateFormat: "2006-01-02"
  myDateFormat : "2006-02-01"
``` 
Then you can generate current date
`${date::now::simpleDateFormat}`
The last parameter is the keyword that you added to config file. You can define multiple date formay to config file and user all of them. Check examples for more info.

!!! You need to use go time format time layout. You can check for more info
[Date format]([quora.com/profile/Ashish-Kulkarni-100](https://gosamples.dev/date-format-yyyy-mm-dd/#:~:text=%F0%9F%93%85%20YYYY-MM-DD%20date%20format%20in%20Go&text=To%2))


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

