migration_up:
	migrate -path migrations -database "postgresql://acltesting:123456@localhost:5432/acltesting?sslmode=disable" -verbose up

migration_down:
	migrate -path migrations -database "postgresql://acltesting:123456@localhost:5432/acltesting?sslmode=disable" -verbose down

migration_fix:
	migrate -path migrations -database "postgresql://acltesting:123456@localhost:5432/acltesting?sslmode=disable" force 1
	 
