use std::fs::File;
use std::io::BufReader;

use serde::{Deserialize, Serialize};

#[derive(Debug, Deserialize, Serialize)]
pub struct Config {
    pub discord_token: String,
    pub ladder_mode: String,
    pub openai_key: String,
    pub mongo_db: String,
    pub mongo_admin: String,
    pub mongo_pass: String,
    pub mongo_uri: String,
    pub mongo_collection_name: String,
}

// function that reads a json file and returns a Config struct
pub fn read_config(path: &str) -> Result<Config, Box<dyn std::error::Error>> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let config = serde_yaml::from_reader(reader)?;

    Ok(config)
}
