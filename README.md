# monitw
Monitoring tool logically consists of three parts:
1.Database containing hosts, types of metrics, metrics, time series of metrics,  authorized users, passwords.

2.Service which in background mode is collecting values of metrics and recording them to db. Implemented using Golang. Source code is in repository monit.

3.Web application which allows: 
	3.1.CRUD operations for hosts, types of metrics, metrics. 
	3.2.Monitor current state of servers metrics and view the history of metrc values.
	
	Backend is implemented as Golang microservices, Source code in repository monitw. 
	Fronted is implemented using React. Source code in repository monitf.
	
	

	
	
