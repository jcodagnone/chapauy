const http = require('http')
const path = require('path')
const NextServer = require('next/dist/server/next-server').default

process.env.NODE_ENV = 'production'
process.chdir(__dirname)

const currentPort = parseInt(process.env.PORT, 10) || 3000
const hostname = process.env.HOSTNAME || 'localhost'

const server = http.createServer(async (req, res) => {
    try {
        // Monkey-patch setHeader to strip x-nextjs-* headers
        const originalSetHeader = res.setHeader;
        res.setHeader = function (key, value) {
            if (typeof key === 'string' && key.toLowerCase().startsWith('x-nextjs-')) {
                return;
            }
            return originalSetHeader.apply(this, arguments);
        };

        await handler(req, res)
    } catch (err) {
        console.error(err)
        res.statusCode = 500
        res.end('internal server error')
    }
})

// Initialize the Next.js server compatible with standalone mode
// This replicates logic from the generated server.js
let handler

async function start() {
    const nextConfig = require('./package.json') // Standalone output puts package.json here? No, let's verify.
    // Actually, standalone relies on .next being present.

    const app = new NextServer({
        hostname,
        port: currentPort,
        dir: path.join(__dirname),
        dev: false,
        conf: {
            ...require('./.next/required-server-files.json').config,
        },
        customServer: false,
    })

    handler = app.getRequestHandler()

    await app.prepare()
    server.listen(currentPort, (err) => {
        if (err) throw err
        console.log(`> Ready on http://${hostname}:${currentPort}`)
    })
}

start()
