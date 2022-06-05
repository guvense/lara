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
Lara is a mock server. You just define your mock server rules as json files and let lara create a server for you. We called this json files as mock spesifications.


# Features
You can check [examples](example) for usage of features.

## Handler
 ### Rest Handler
Rest Handler handles your rest request and return response as you defined.
- Regex Match     
Lara supports regex for matching request with mocks thanks to gorilla mux library. If the regex matched path variable or request paramaters, lara will match request to defined mocks and will return response that belongs to matched request.
- Response Delay     
You can also set delay time to similate real life scenarios. You only add after field to response or callback objects.

## Callback
  ### Rest Callback
- You can define callback method for each request.First lara will return response according to spec then lara invoke your callback action. As mentioned, you can define a delay time for callbacks as well.


## Dynamic Value
Thanks to dynamic value feature, you can generate unique values for each request. Therefore you can get generic response or send generic request to callback services. There are different libraries for these fields. You only put these scripts to your mock spesification.

### String Library
- Random String    
It will generate ten lenght random string.You can adjust lenght of string by changing last paramater.    
`${str::random::10} `
- Uuid String    
It will generate uuid   
`${str::uuid} `
- Random String From Regex   
You can also generate random string by defining your own regex format. First you need you define your regular expression to config.yml file as below. You are not allowed to define these regular expression to mock spesification because of parsing rules.
Otherwise, your spesification will behave unexpected.
``` config.yml
regex:
  email: "^[a-z]{5,10}@[a-z]{5,10}\\.(com|net|org)$"
  number: "/\d+"
``` 
Then you can use this expression with regex keyword. 10 is the lenght of string. If you provide lenght of string in your regular expression you don't need to pass string lenght as paramater. Check example for more usages.
`${str::regex::email::10} `

### Number Library
- Random Number      
Generating a random number between -10 and 1    
`${number::generate::-10::1}`


### Date Library
- Current Date from format   
You can generate current date from format    
First you need to define your date format to config file as below.
If you are not provided any format the default format will be;
`01-02-2006 -> MM-DD-YYYY`

``` 
date:
  simpleDateFormat: "2006-01-02"
  myDateFormat : "2006-02-01"
``` 
Then you can generate current date
`${date::now::simpleDateFormat}`
The last parameter is the keyword that you added to config file. You can define multiple date formay to config file and user all of them. You are not allowed to define format to mock spesification. Check examples for more info.

!!! You need to use go time format time layout. You can check for more info
[Date format]([quora.com/profile/Ashish-Kulkarni-100](https://gosamples.dev/date-format-yyyy-mm-dd/#:~:text=%F0%9F%93%85%20YYYY-MM-DD%20date%20format%20in%20Go&text=To%2))


### Value keeper 
You can extract values from request and response with using value keeper.     
For instance this script will retrieve user from request query parameter. For more you can check examples.     
`${value::request::queryparams::userId}`


## Authorization Interceptor
If you define callback method and this callback will send request to a server. A server is protected with an authorization method. You can define your authorization credential to config yaml file and use it as value of token-generator. Before sending request to callback server, lara will first fetch token and use this token while fetching callback service.       
First you need to define your authorization service credential as below.      
As an example keycloak will our token server name. You can give any name that you want instead of keycloak.  
Lara supports Password and Client Credentials grant types. Check example config yml for more.
``` 
type : Password-Credential -> Password Credentials Grant Type
type :  Client-Credential  -> Client Credentials Grant Type


``` 

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


## Watcher 
If you want to update server when you update mock file, you can basically pass watcher flag to application.     
There two way to initialize lara with watcher   
First pass wather flag to application   
`--watcher`   
Second add watcher field to config file   
`watcher: true`     


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


## Contribution Guide
You just create a pull request with description!    

### Next step

1- Documentation  
2- Test Application according to doc      
3- Bug Fixing     
4- Adding basic test cases    
5 -First Release!!!!!      


