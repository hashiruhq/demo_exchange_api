Hashiru Demo Exchange API
=========================

A demo API for connecting to the [Hashiru Trade Engine](https://gum.co/rNsKn).

## Usage

1. Subscribe to [Hashiru Trade Engine](https://gum.co/rNsKn) to download a copy of the engine.
2. Download the code from this repo in a demo_exchange_api folder
3. Add the following to the `docker-compose.yml` file to build it and optionally map some directories to restart on file changes:

    ```yaml
    engine_api:
      build:
        context: ../demo_exchange_api
        dockerfile: Dockerfile.dev
      image: engine_api:local
      volumes:
        - ../demo_exchange_api/server:/build/demo_api/server
      ports:
        - 3080:80
    ```

4. Start the engine using the docker up command from the trade engine folder: `docker-compose -p starter up -d --build`
5. Make API calls to `POST/DELETE http://localhost:3080/orders/btcusdt` for create/cancel an order
6. Check out the logs to see the results
