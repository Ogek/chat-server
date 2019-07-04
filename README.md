# Go Chat Server
Simple chat server with WebSocket in Go

## Usage
From client do all your requests sending an object with key `type` to request various actions to the server and `payload` to send the content.

### `type:"login"`
Add the user to the chat.
`payload` keys:
- `name`: Name of the user


### `type:"message"`
Send message from user.
`payload` keys:
- `text`: Body of message


# english: sorry I had to, :D
User leaves the chat automatically on closed connection 

## Examples

### Login
```javascript
ws.addEventListener("open", e => {
    var name = prompt("Chose username", "");
    var data = {
        type: "login",
        payload: {
            name
        }
    }
    ws.send(JSON.stringify(data))
});
```

### Send message
```javascript
form.addEventListener("submit", e => {
    e.preventDefault()
    var msg = document.getElementById("msg").value
    var data = {
        type: "message",
        payload: {
            text: msg
        }
    }
    ws.send(JSON.stringify(data))
    e.target.reset()
});
```

### Receive message
```javascript
ws.addEventListener("message", e => {
    var d = JSON.parse(e.data)
    const {type, payload} = d
    switch(type) {
        case "message": //or "admin_message" for admin messages
            var p = document.createElement("p");
            p.textContent = `{payload.user.name}:{payload.text}`;
            chatContainer.append(p);    
            break
        ...
    }
});
```


