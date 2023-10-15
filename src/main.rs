use std::env;
use std::process;

mod config;
mod discordbot;

#[tokio::main]
async fn main() {

	// Read the initial config file.
	// Note: When run from a container, this will need to be mounted in as a secret.
    let default_file = String::from("config.yml");
	
    let args: Vec<String> = env::args().collect();
    let config_path = args.get(1).unwrap_or(&default_file);

    let conf = match config::read_config(config_path) {
        Ok(c) => c,
        Err(e) => {
            eprintln!("Error reading config file: {}", e);
            process::exit(1);
        }
    };

    discordbot::run(conf).await;
    
    println!("Exiting...");
}
