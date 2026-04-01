use serde::{Serialize, Deserialize};
use std::io::{self, BufRead};
use serde_json::json;

#[derive(Serialize, Deserialize)]
struct Message {
    src: String,
    dest: String,
    body: serde_json::Value,
}

fn main() {
    let stdin = io::stdin().lock();
    let mut node_id = String::new();
    for line in stdin.lines() {
        let line = line.expect("failed to read line");
        let msg: Message = serde_json::from_str(&line).expect("failed to parse");

        let reply_body = match msg.body["type"].as_str() {
            Some("init") => {
                node_id = msg.body["node_id"].as_str().unwrap().to_string();
                json!({
                    "type": "init_ok",
                    "in_reply_to": msg.body["msg_id"]
                })
            }
            Some("echo") => {
                json!({
                    "type": "echo_ok",
                    "in_reply_to": msg.body["msg_id"],
                    "echo": msg.body["echo"]
                })
            }
            _ => {
                eprintln!("unknown message type");
                continue;
            }
        };

        let reply = Message {
            src: node_id.clone(),
            dest: msg.src,
            body: reply_body,
        };
        println!("{}", serde_json::to_string(&reply).unwrap());
    }
}
