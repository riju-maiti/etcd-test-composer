use std::thread;
use std::time::Duration;
use etcd_client::*;
use tokio;

#[tokio::main]
async fn main() -> Result<(), Error> {

    antithesis_init()

    let duration = Duration::new(15, 0);
    thread::sleep(duration);

    let mut client = Client::connect(["http://etcd0:2379"], None).await?;

    let options = GetOptions::default().with_all_keys();

    // get keys
    let response = client.get("", Some(options)).await?;

    // count
    let count = response.kvs().len();

    println!("Number of entries in etcd: {}", count);

    // add sdk assertion that this number is always less than or equal to 6, since that is our key space

    Ok(())
}