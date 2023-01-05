import { client, connection } from 'websocket'

const eventsuburl = "wss://eventsub-beta.wss.twitch.tv/ws"

const ws = new client();
const channel = process.env.TWITCH_CHANNEL;  // Replace with your channel.
const account = process.env.TWITCH_ACCOUNT;   // Replace with the account the bot runs as
const password = process.env.TWITCH_PASSWORD;

const connectFunc = (connection: connection) => {
    console.log('WebSocket Client Connected');

    connection.on('message', (msg) => {
        if (msg.type === 'utf8') {
            let rawMsg = msg.utf8Data.trimEnd();
            let objmsg: any = null
            try {
                objmsg = JSON.parse(rawMsg)
            } catch {
                return;
            }

            switch (objmsg.metadata.message_type) {
                case "session_welcome":
                    console.log("Welcome", objmsg.payload.session.status)
                    break;
                default:
                    console.log("missing", objmsg);
                    break;
            }
        }
    });
}

ws.on('connectFailed', (error) => {
    console.log('Connect Error: ' + error.toString());
});

ws.on('connect', connectFunc);