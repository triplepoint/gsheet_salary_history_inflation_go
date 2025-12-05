# TL;DR

If the token is expired:

```shell
rm token.json
```

then run the script:

```shell
./gsheet_salary_history_inflation_go
```

Open the generated link in a browser, follow the auth, and then watch the localhost URL it sends you to fail. Something like:

```
http://localhost/?state=state-token&code=4/0Ab32jghdfjghfghdghdjgdshgdjgdG09VuJRhpVhqif3csMlG24KVN6j7xGyQ&scope=https://www.googleapis.com/auth/spreadsheets
```

Copy out the value from `code=` in this case "4/0Ab32jghdfjghfghdghdjgdshgdjgdG09VuJRhpVhqif3csMlG24KVN6j7xGyQ" and paste it into the terminal where the script is running.

You can run the script as many times as you want until the token expires and you have to redo the above.
