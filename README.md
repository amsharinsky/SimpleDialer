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

#

  - [Install](#install)
    - [DATABASE](#DATABASE)
      - [Make dialer tables](#Make-dialer-tables)
	



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


### DATABASE

#### Make dialer tables

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
```typescript
CREATE TABLE <Scheme>.dialer_params (
	project_id varchar NULL,
	lines int8 NULL,
	call_time int8 NOT NULL DEFAULT 20,
	case_limit int8 NULL,
	sort varchar NULL,
	id bigserial NOT NULL,
	"type" varchar NULL,
	exten varchar NULL,
	context varchar NOT NULL DEFAULT 'default'::character varying
);
```

```typescript
CREATE TABLE <Scheme>.dialer_stat (
	id int4 NOT NULL,
	case_name varchar NULL,
	project_id varchar NULL,
	phone_number varchar NULL,
	dial_count int4 NOT NULL DEFAULT 0,
	ended timestamp NULL,
	state varchar NULL
);
```
* For Mariadb

```typescript
CREATE TABLE `<Scheme>.dialer_clients` (
  `case_name` varchar(100) NOT NULL,
  `project_id` varchar(100) NOT NULL,
  `phone_number` varchar(100) NOT NULL,
  `priority` int(11) DEFAULT 1,
  `utc` varchar(100) DEFAULT NULL,
  `used` tinyint(1) NOT NULL DEFAULT 0,
  `recall` tinyint(1) DEFAULT 0,
  `allowed_start` varchar(100) NOT NULL DEFAULT '00:00:00',
  `allowed_stop` varchar(100) NOT NULL DEFAULT '00:00:00',
  `checked` tinyint(1) NOT NULL DEFAULT 0,
  `recall_period` varchar(100) DEFAULT NULL,
  `chime_time` timestamp NULL DEFAULT NULL,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8
```

```typescript
CREATE TABLE `<Scheme>.dialer_params` (
  `project_id` varchar(100) COLLATE utf8_estonian_ci NOT NULL,
  `lines` int(11) NOT NULL DEFAULT 0,
  `call_time` int(11) NOT NULL DEFAULT 20000,
  `case_limit` int(11) NOT NULL DEFAULT 500,
  `sort` varchar(100) COLLATE utf8_estonian_ci NOT NULL DEFAULT '2:1',
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type` varchar(100) COLLATE utf8_estonian_ci NOT NULL DEFAULT 'progressive',
  `exten` varchar(100) COLLATE utf8_estonian_ci NOT NULL,
  `context` varchar(100) COLLATE utf8_estonian_ci NOT NULL DEFAULT 'default',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8 COLLATE=utf8_estonian_ci
```

```typescript
CREATE TABLE `<Scheme>.dialer_stat` (
  `id` int(11) NOT NULL,
  `case_name` varchar(100) DEFAULT NULL,
  `project_id` varchar(100) DEFAULT NULL,
  `phone_number` varchar(100) DEFAULT NULL,
  `dial_count` int(11) NOT NULL DEFAULT 0,
  `ended` timestamp NULL DEFAULT NULL,
  `state` varchar(100) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8
```

#### Create dialer cases

