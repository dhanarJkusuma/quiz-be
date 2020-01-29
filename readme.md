This project was built using Golang Programming Language, Redis, MySQL, and Socket.IO

## Requirements
<ul>
    <li>Golang</li>
    <li>MySQL</li>
    <li>Redis</li>
    <li>Dep Go Dependency Management Tool</li>
    <li>Node.JS (For Front End Application)</li>
</ul>

## How to start?
1. Clone the repository
```
git clone https://github.com/dhanarJkusuma/quiz-be.git quiz
```
2. Install Dep, here is the link [Dep Golang](https://github.com/golang/dep)
3. Run Redis and MySQL Server
4. Setting your config in `env.json`
    ```json
    {
      "base_url": "localhost:8000", 
      "quiz": {
        "number_of_question": 2,
        "ready_count_down": 3, 
        "count_down": 3,
        "template_path": "C://Users/BN001706734/go/src/github.com/dhanarJkusuma/quiz/templates" 
      },
      "database": {
        "address":"127.0.0.1:3306",
        "name": "quiz_2",
        "user": "root",
        "password": ""
      },
      "redis": {
        "address": "127.0.0.1:6379",
        "password": ""
      },
      "jwt": {
        "secret_key": "A1gW8lAPwN"
      }
    }
   ```
   <ul>
       <li>`base_url`:  your backend base url.</li>
       <li>`number_of_question`: total number of question that asked for user.</li>
       <li>`ready_count_down`: countdown before quiz is started </li>
       <li>`count_down`: countdown during quiz </li>
       <li>`template_path`: html template path</li>
       <li>`secret_key`: get the random string for security purpose</li>
   </ul>
5. Go to root project directory and run script: `dep ensure -v`
6. To start the project, run: `go run *.go`

## Available Route
<ul>
    <li>Admin Dashboard Login Page: <b>http://localhost:8000/admin/login</b> </li>
</ul>

## Note
* Run [Quiz-Front-End](https://github.com/dhanarJkusuma/quiz-fe) project, to start quiz application.
* If you want to run Redis for Windows, here I'm using this app [Redis For Windows](https://github.com/microsoftarchive/redis/releases) 