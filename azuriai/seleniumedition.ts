import http from "http";
import { Builder, By, Key, until, WebElement } from "selenium-webdriver";
import { Options } from "selenium-webdriver/firefox";

const options = new Options().setProfile("../profile");
options.setBinary(process.env.FFBIN)

const driver = new Builder()
  .forBrowser("firefox")
  .setFirefoxOptions(options)
  .build();

driver.get("https://chat.openai.com/chat/c5a5bee5-820f-4a54-888c-47bd8765e968");
// driver.get("https://platform.openai.com/playground?mode=chat")

const hostname = "127.0.0.1";
const port = 3000;

let isBusy = false;

let currentResult = "";

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
          res.end(JSON.stringify({ result }));
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

async function processAzuriAI(input: string): Promise<string> {
  const textArea = await driver.findElement(By.css("textarea"));
  await textArea.sendKeys(input + Key.ENTER);
  await driver.wait(
    until.elementLocated(
      By.xpath("//*[contains(text(), 'Regenerate response')]")
    ),
    25000
  );
  const lastPTag: WebElement = await driver.executeScript(`
    const pTags = document.getElementsByTagName('p');
    return pTags[pTags.length - 2];
  `);
  const newText = await lastPTag.getAttribute("innerText");

  return Promise.resolve(newText);
}

// enter prompt here document.querySelector(".chat-pg-instructions>div>textarea")
// toggle message to assistant document.querySelector("span.chat-message-role-text")
// submit button document.querySelector(".pg-content-footer>button")