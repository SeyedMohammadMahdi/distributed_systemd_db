# Go-DB
this project is a simple key-value database. 
- it uses `Gin framework` as a backbone for the web server which is a light weight and fast.
- for the persistence and data storage there is a module named `Badger` it helps in storing and retrieving stored data in an efficient way

## project overview 

this project contains 3 routes:

- `GET /objects` 
	- which is used to get all the stored key value pairs

- `GET /objects/{key}` 
	- that is used for getting a specific value

- `PUT /objects` 
	- that is used to store a new key-value pair in the database or update an existing value

## setup

for running the project the only thing you need to do is having docker and docker-compose on you system and follow the instructions below:


building the docker image:
```bash
docker compose build
```

running the container:
```bash
docker compose up
```

and there you go. you have a simple database.

### usage

adding new data:
```bash
curl -X PUT http://127.0.0.1:8080/objects \
  -H "Content-Type: application/json" \
  -d '{
    "key": "user:1",
    "value": {
      "name": "Amin Alavi zadeh",
      "age": 23,
      "email": "a.alavi@fum.ir"
    }
  }'
```

the response is only the status code which is 200 for success and 400 for failure

reading a specific data:
```bash
curl http://127.0.0.1:8080/objects/user:1
```

if the request is fine you get a JSON response like the one below otherwise you get 404 or 400
```bash
HTTP/1.1 404 Not Found
Content-Type: text/plain
Date: Fri, 31 Oct 2025 13:01:13 GMT
Content-Length: 18
```

```json
{
  "age": 23,
  "email": "a.alavi@fum.ir",
  "name": "Amin Alavi zadeh"
}
```

getting all datas:
```bash
curl http://127.0.0.1:8080/objects
```

if there is anything in the data base it will return  a json which is a list of objects containing 2 fields, key and value, but if there is nothign there then you get that famous 404.

```bash
HTTP/1.1 404 Not Found
Content-Type: text/plain
Date: Fri, 31 Oct 2025 13:08:26 GMT
Content-Length: 18
```

```json
[
  {
    "key": "user:1",
    "value": {
      "age": 23,
      "email": "a.alavi@fum.ir",
      "name": "Amin Alavi zadeh"
    }
  },
  {
    "key": "user:2",
    "value": {
      "age": 24,
      "email": "a.alavi@fum.ir",
      "name": "mohammad mahdi"
    }
  }
]
```


### adding data to the database
<img width="1366" height="612" alt="image" src="https://github.com/user-attachments/assets/b8e6435c-752d-46ae-9b43-4ef86f5b12d6" />


### getting all datas stored in database
<img width="1366" height="608" alt="image" src="https://github.com/user-attachments/assets/dbd2c813-f54b-4a52-92c5-16fa89642873" />


### getting a single data
<img width="1361" height="594" alt="image" src="https://github.com/user-attachments/assets/fe673237-5b21-404d-a56a-516cd3fcfc2f" />

