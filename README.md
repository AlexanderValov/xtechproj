# xtechproj

### Running:

- command: make build & make run 
- NOTE: need to run docker first
- port: 8000 

### Endpoints

- /api/btcusdt - GET: return last data for BTC
- /api/btcusdt - POST: return history for BTC
<br><br>
- /api/currencies - GET: return last data for Fiat
- /api/currencies - POST: return history for Fiat
<br><br>
- /api/latest - GET: returns BTC/Fiat

### Filters for POST requests:

- limit (~?limit=5)
- offset (~?offset=5)
- order_by: (~order_by=-value)
    - for BTC:
        - value/-value;
        - created_at/-created_at;
        - latest/-latest
  - for Fiat:
      - created_at/-created_at;
      - latest/-latest

example: /api/btcusdt?limit=10&offset=10&order_by=created_at
