package model

//go:generate goctl model mysql ddl -c -src user.sql -dir .

//go:generate goctl model mysql datasource -url="$datasource" -table="user" -c -dir .
