
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


## Dynamic Value
Thanks to the dynamic value feature, you can generate unique values for each request. Therefore you can get generic response or send a generic requests to callback services. There are different libraries for these fields. You only put these scripts to your mock specification.         

### String Library
- Random String    
It will generate ten length random string. You can adjust the length of the string by changing the last parameter.     
`${str::random::10} `
- Uuid String    
It will generate uuid   
`${str::uuid} `
- Random String From Regex   
You can also generate random strings by defining your own regex format. First, you need you to define your regular expression to config.yml file as below. You are not allowed to define this regular expression to mock specifications because of parsing rules.       
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
Then you can generate current date with the script     
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
If you define callback method and this callback will send request to a server. In some cases, that server is protected with an authorization method. You can define your authorization credential to config yaml file and use it as value of token-generator into callback specification. Before sending request to the callback server, Lara will first fetch token and use this token while fetching the callback service.            
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


## Usage 

### Docker
- Passing arguments 

```
docker run -p 8899:8899 -v "$PWD/mocks.json:/mocks.json" -v "$PWD/config.yml:/config.yml"  guvense/lara  --config /config.yml --mocks /mocks.json
```


## License
MIT License, see [LICENSE](https://github.com/guvense/lara/blob/main/LICENSE)


Inspired by  [killgrave](https://github.com/friendsofgo/killgrave)
