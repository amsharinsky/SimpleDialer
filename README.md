# SimpleDialer
This is alfa version.Simple dialer is a system for automatic dialing of subscribers using an asterisk.
- Used only by AMI
- Simple configuration
- Progressive mode
- Autoinforming mode

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

  - [Install](#install)
    - [Database](#Database)
      - [Create dialer tables](#Create-dialer-tables)
      	- [Сolumn description](#Сolumn-description)
      - [Create dialer cases](#Create-dialer-cases)
  - [Dialer management](#Dialer-management)  

### Install
	
- Create tables in database
- Configure asterisk and dialer configuration files
- Create cases for calling
- Send Get query for start

### Dialer management
- A get request is used to control the dialer
- Parameters used: 
   action  and projectid where action can be start or stop

Example:http://127.0.0.1:8080/?action=start&projectid=test

 ### Configuration files
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


### Database

#### Create dialer tables

* For Postgresql
   
```typescript
  CREATE TABLE <Scheme>.dialer_clients (
  
	case_name varchar NOT NULL,
	project_id varchar NOT NULL,
	phone_number varchar NOT NULL,
	priority int8 NOT NULL DEFAULT 1,
	utc varchar NULL,
	used bool NOT NULL DEFAULT false,
	call_back bool NOT NULL DEFAULT false,
	allowed_start varchar NOT NULL DEFAULT '00:00:00'::character varying,
	allowed_stop varchar NOT NULL DEFAULT '00:00:00'::character varying,
	checked bool NOT NULL DEFAULT false,
	call_back_period varchar NULL,
	deferred_time timestamp NULL,
	deferred_done bool NOT NULL DEFAULT false,
	created timestamp NOT NULL DEFAULT now(),
	id serial NOT NULL
  
);
```
```typescript
CREATE TABLE <Scheme>.dialer_params (
	project_id varchar NULL,
	lines int8 NULL,
	dial_time int8 NOT NULL DEFAULT 20,
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
  `call_back` tinyint(1) DEFAULT 0,
  `allowed_start` varchar(100) NOT NULL DEFAULT '00:00:00',
  `allowed_stop` varchar(100) NOT NULL DEFAULT '00:00:00',
  `checked` tinyint(1) NOT NULL DEFAULT 0,
  `deferred_done` tinyint(1) NOT NULL DEFAULT 0,
  `call_back_period` varchar(100) DEFAULT NULL,
  `deferred_time` timestamp NULL DEFAULT NULL,
  `created` timestamp NOT NULL DEFAULT current_timestamp(),
  `id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8
```

```typescript
CREATE TABLE `<Scheme>.dialer_params` (
  `project_id` varchar(100) COLLATE utf8_estonian_ci NOT NULL,
  `lines` int(11) NOT NULL DEFAULT 0,
  `dial_time` int(11) NOT NULL DEFAULT 20000,
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

### Сolumn description

	-dialer_clients
	
| Column name             | Column description          
| ----------------------- | ----------------------- |
| case_name               | Case name.Сan be anything|
| project_id              | Unique project ID       |
| phone_number            | Subscriber's phone number      |       
| priority                | Priority of the case. 1-maximum priority. The higher the priority, the sooner the case will call           |   
| utc                     | Subscriber's utc. By default, the time from the server is taken   | 
| used			  | System field.If it is set true, then the case is called|	
| call_back		  | Allow call back case |
| allowed_start		  | Beginning of the allowed call time of the case|			
| allowed_stop		  | End of the allowed time for the case call |	
| checked		  | System field |	
| call_back_period	  | Call back period in minutes. Example: 1,1.There will be two dialing attempts with an interval of one minute.|
| deferred_time		  | System field |
| deferred_done		  | Is the callback of the deferred case completed|	
| created		  | Case creation time| 
| id			  | System field |

	-dialer_params
	
| Column name             | Column description          
| ----------------------- | ----------------------- |
| project_id              | Unique project ID        |
| lines           	  | Maximum number of lines for simultaneous calls|       
| dial_time               | Call time to subscriber.in milliseconds |            |   
| case_limit              | The number of cases taken for dialing in memory | 
| sort			  | Sorting cases. Consists of two parameters: the first parameter: 1-priority, 2-date of creation, 3-UTC.The second parameter: 1-DESC, 2-ASC. Example: 2:1 |
| type			  | Dialer mode. progressive or autoinforming.Example: progressive |
| exten			  | Project (exten) Number in extensions.conf|
| context	          | Context in extensions.conf |

	-dialer_stat

| Column name             | Column description          
| ----------------------- | ----------------------- |
| id            	  | System field.Unique      |
| case_name           	  | Case name.Сan be anything|       
| project_id              | Unique project ID |            |   
| phone_number            | Subscriber's phone number | 
| dial_count		  | Number of calls already made | 
| ended			  | Time of the end of the case call |
| state			  | Call status |





#### Add dialer cases and params(Example)

INSERT INTO nc.dialer_clients
(case_name, project_id, phone_number, priority, utc,recall, allowed_start, allowed_stop, recall_period)
VALUES('Name', 'test', '1111', 1 , '+03.00h',true, '10:00:00', '18:00:00','1,1' );

INSERT INTO nc.dialer_params
(project_id, `lines`, call_time, case_limit, `type`, exten, context)
VALUES('test', 0, 20000, 500,  'progressive', '2222', 'default');




