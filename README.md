# TTT-IM

A real-time messaging and instant communication module for the TTT (Tic-Tac-Toe) server application.

## Features

- Real-time message delivery
- User presence tracking
- Connection management
- Message history

## Installation

```bash
npm install ttt-im
```

## Usage

```javascript
const TTTIm = require('ttt-im');

const im = new TTTIm({
    port: 3000
});

im.start();
```

## API Documentation

See [API.md](./API.md) for detailed documentation.

## Contributing

Contributions are welcome. Please open an issue or submit a pull request.

## License

MIT