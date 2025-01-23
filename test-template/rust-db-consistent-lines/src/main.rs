use std::thread;
use std::time::Duration;
use etcd_client::*;
use tokio;
use serde_json::{json};
use antithesis_sdk::{assert_always, antithesis_init};

#[tokio::main]
async fn main() -> Result<(), Error> {

    antithesis_init();

    // When finally commands are run, the fault injector is turned off. Sleeping 15 seconds allows the system to recover.
    let duration = Duration::new(15, 0);
    thread::sleep(duration);

    let mut client = Client::connect(["http://etcd0:2379"], None).await?;

    let options = GetOptions::default().with_all_keys();

    // Get all the keys that are in the etcd cluster
    let response = client.get("", Some(options)).await?;

    // Count the number of keys that we received
    let count = response.kvs().len();

    println!("Number of entries in etcd: {}", count);

    // Always assertion that the key space is less than or equal to 6. We only use 6 keys in the generate_traffic script.
    // Note: even though we have add 3 keys from the eventually_health_check script, an eventually command and a finally command
    // will never be ran in the same timeline. That eliminates the key space from being 9 whenever we run this rust script.
    let details = json!({"key_space_size": count});
    assert_always!(count <= 6, "Key space remains bounded", &details);

    Ok(())
}