use poise::serenity_prelude as serenity;


use crate::config::Config;
//use crate::rankingdata::RankingData;


pub struct DiscordBot {
    //pub client: serenity::Client,
    //ranking_data: RankingData,
    pub db_client: mongodb::Client,
}


// ladder command with subcommands
#[poise::command(
    slash_command,
    subcommands(
        //"help",
        "init",
        "delete_tournament",
        "register",
        "unregister",
        "challenge",
        "result",
        "cancel",
        "forfeit",
        "r#move",
        "standings",
        "history",
        "printraw",
        "set",
    ),
)]
async fn ladder(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Help is on the way!
/* #[poise::command(
	slash_command,
	help_text = "",
)]
async fn help(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
} */

// Initialize a 1v1 ranking tournament.
#[poise::command(
	slash_command,
)]
async fn init(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Delete a 1v1 ranking tournament.
#[poise::command(
	slash_command,
)]
async fn delete_tournament(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Register a user to a 1v1 ranking tournament.
#[poise::command(
	slash_command,
)]
async fn register(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "Specify user, admin only"]	user: Option<serenity::User>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Unregister a user from a 1v1 ranking tournament.
#[poise::command(
	slash_command,
)]
async fn unregister(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "Specify user, admin only"]	user: Option<serenity::User>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Challenge a user to a 1v1 match.
#[poise::command(
	slash_command,
)]
async fn challenge(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "Specify user"]	user: serenity::User,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// "Report the result of a challenge.
#[poise::command(
	slash_command,
)]
async fn result(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "The result of the challenge."]	result: String, //TODO: Enum won, lost, forfeit, cancel
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Cancel a challenge.
#[poise::command(
	slash_command,
)]
async fn cancel(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Forfeit a challenge.
#[poise::command(
	slash_command,
)]
async fn forfeit(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Move a user to a different position in the ladder, admin only.
#[poise::command(
	slash_command,
)]
async fn r#move(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "Specify user"]	user: serenity::User,
	#[description = "New position"]	position: u32,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Get the current standings.
#[poise::command(
	slash_command,
	aliases("rank", "print"),
)]
async fn standings(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Get the recent history of matches.
#[poise::command(
	slash_command,
)]
async fn history(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "history length"] length: Option<u32>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

// Print the raw data for the channel.
#[poise::command(
	slash_command,
)]
async fn printraw(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    //ctx.data().channels
    Ok(())
}

// Set a value in the ranking data.
#[poise::command(
	slash_command,
)]
async fn set(
    ctx: poise::Context<'_, DiscordBot, Box<dyn std::error::Error + Send + Sync>>,
	#[description = "user or system"] level: String, //TODO: Enum user, system
	#[description = "key"]	key: String, //TODO: Enum notes, status, mode, timeout, admin_add, admin_remove
	#[description = "value"]	value: String, //TODO: Enum notes, status, mode, timeout, admin_add, admin_remove
	//TODO these might need to be subcommands 
) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let u = ctx.author();
    let response = format!("Hi {}! You made it do a thing!", u.name);
    ctx.say(response).await?;
    Ok(())
}

pub async fn run(config: Config) {
    // Create context data for the bot.
    let data = DiscordBot {
        db_client: mongodb::Client::with_uri_str(&config.mongo_uri).await.unwrap(),
    };

    // Build the poise framework.
    let framework = poise::Framework::builder()
        .options(poise::FrameworkOptions {
            commands: vec![ladder()],
            ..Default::default()
        })
        .token(config.discord_token.clone())
        .intents(serenity::GatewayIntents::non_privileged())
        .setup(|ctx, _ready, framework| {
            Box::pin(async move {
                poise::builtins::register_globally(ctx, &framework.options().commands).await?;
                Ok(data)
            })
        });

    // Run the bot or die in glory... or something.
    framework.run().await.unwrap();
}