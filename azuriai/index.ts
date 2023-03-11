import axios from "axios";
import * as http from "http";
import * as fs from "fs";
import * as dotenv from "dotenv";

if (fs.existsSync(".env")) {
  dotenv.config();
}

const hostname = "0.0.0.0";
const port = 3000;

let currentResult = "";

let isBusy = false;

const server = http.createServer(async (req, res) => {
  switch (req.url) {
    case "/azuriai":
      if (req.method !== "POST") {
        res.statusCode = 405;
        res.setHeader("Allow", "POST");
        res.setHeader("Content-Type", "text/plain");
        res.end("Method Not Allowed");
        return;
      }
      let body = "";
      req.on("data", (chunk) => {
        body += chunk.toString();
      });
      req.on("end", async () => {
        try {
          const data = JSON.parse(body);
          console.log("got body for azuriai", data);
          if (!data || !data.content) {
            res.statusCode = 400;
            res.setHeader("Content-Type", "text/plain");
            res.end("Bad Request: missing content property");
            return;
          }
          if (isBusy) {
            res.statusCode = 400;
            res.setHeader("Content-Type", "application/json");
            res.end('{"content":"Server busy"}');
            return;
          }
          isBusy = true;
          // Perform some asynchronous operation
          const result = await processAzuriAI(data.content);
          isBusy = false;
          if (result === currentResult) {
            res.statusCode = 400;
            res.setHeader("Content-Type", "application/json");
            res.end(JSON.stringify({ result: "AI failed..." }));
            return;
          }
          currentResult = result;
          res.statusCode = 200;
          res.setHeader("Content-Type", "application/json");
          res.end(JSON.stringify({ result: result.replaceAll("*", "") }));
        } catch (err: any) {
          isBusy = false;
          res.statusCode = 400;
          res.setHeader("Content-Type", "text/plain");
          res.end(`Bad Request: ${err.message}`);
        }
      });
      break;
    default:
      res.statusCode = 200;
      res.setHeader("Content-Type", "text/plain");
      res.end("Pog!");
      break;
  }
});

server.listen(port, hostname, () => {
  console.log(`Server running at http://${hostname}:${port}/`);
});

interface ChatState {
  model: string;
  messages: { role: string; content: string }[];
  max_tokens: number;
  temperature: number;
}

interface ChatResponse {
  choices: {
    message: { role: string; content: string };
    finish_reason: string;
    index: number;
  }[];
}

async function processAzuriAI(content: string): Promise<string> {
  const state: ChatState = JSON.parse(
    fs.readFileSync("state/state.json").toString()
  );
  state.messages.push({ role: "user", content });

  const { data, status, statusText } = await axios.post<ChatResponse>(
    "https://api.openai.com/v1/chat/completions",
    state,
    {
      headers: {
        Authorization: "Bearer " + process.env.OPENAI_APIKEY,
        "Content-Type": "application/json",
      },
    }
  );
  if (
    status != 200 ||
    data.choices.length < 1 ||
    data.choices[0].finish_reason != "stop"
  ) {
    console.log("SOMETHING HAS GONE TERRIBLY WRONG", status, statusText, data);
    Promise.resolve(currentResult);
  }
  // Should not save new state, doesnt work when requests are too big
  // state.messages.push(data.choices[0].message);
  // fs.writeFileSync("state/state.json", JSON.stringify(state, null, 4));

  return Promise.resolve(data.choices[0].message.content);
}
