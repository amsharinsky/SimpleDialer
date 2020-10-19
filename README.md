# SimpleDialer
This is alfa version.Simple dialer is a system for automatic dialing of subscribers using an asterisk.
- Used only by AMI
- Simple configuration
- Progressive mode
- Autoinfo mode

### In future releases
- Predictive mode
- Webphone 
- Realtime statistics and more

### Software requirements

Asterisk:
  - Asterisk 16.X

Databases that can be used:
 - Postegresql 12.X
 - MariaDB 10.3.X

## Usage
 ### 1. Configuration files
    -/config/asteriks.json

| Param name              | Param description             | Type    | Example     |
| ----------------------- | ----------------------- | --------| --------    |
| Ip                      | Asterisk ip address     | String  | 192.168.1.1 | 
| Port                    | Asterisk ami port       | String  | 5038        |
| Timeout                 | Connection timeout      | Integer | 10          |        
| Username                | AMI username            | String  | User        |   
| Password                | AMI username password   | String  | 123qwe      |  
  
   -/config/dbconfig.json 
   
| Param name              | Param description                                                         | Type    | Example     |
| ----------------------- | ------------------------------------------------------------------- | --------| ----------- |
| Driver                  | Driver for connect DB                                               | String  | pgx or mysql| 
| Port                    | DB port                                                             | Integer | 5432        |
| DB_Server               | DB ip address                                                       | String  | 192.168.1.2 |        
| DB_Name                 | DB name                                                             | String  | Asterisk    |   
| Scheme                  | Scheme name                                                         | String  | nc          |
| Username                | Username with full  rights for the scheme                           | String  | nc_username |
| Password                | Username password                                                   | String  | 111222      |
| Conn_Max_Lifetime       | Maximum amount of time a connection may be reused                   | Integer | 5           | 
| Max_Open_Conns          | Maximum number of open connections to the database                 | Integer | 100         |
| Max_Idle_Conns          | Maximum number of connections in the idle connection pool           | Integer | 10          |

 -/config/dialer.json
 
| Param name              | Param description                                    | Type     | Example |
| ----------------------- | ---------------------------------------------------- | -------- |-------- |
| loglevel                | Set loglevel. 1-ERROR.2-ERROR,INFO.3-ERROR,INFO,DEBUG | Integer  | 1       | 
  
  -/config/http.json 

| Param name              | Param description                                    | Type     | Example |
| ----------------------- | ---------------------------------------------------- | -------- |-------- |
| Ip                      | Server ip address where you want to start dialer service| String  | 192.168.1.3|
| Port                    | Port of dialer service                               | String  | 8080     |

### Make dialer tables

* For Postgresql
   
```typescript
  CREATE TABLE <Scheme>.dialer_clients (
  
	case_name varchar NOT NULL,
	project_id varchar NOT NULL,
	phone_number varchar NOT NULL,
	priority int8 NOT NULL DEFAULT 1,
	utc varchar NULL,
	used bool NOT NULL DEFAULT false,
	recall bool NOT NULL DEFAULT false,
	allowed_start varchar NOT NULL DEFAULT '00:00:00'::character varying,
	allowed_stop varchar NOT NULL DEFAULT '00:00:00'::character varying,
	checked bool NOT NULL DEFAULT false,
	recall_period varchar NULL,
	chime_time timestamp NULL,
	created timestamp NOT NULL DEFAULT now(),
	id serial NOT NULL
  
);
```



