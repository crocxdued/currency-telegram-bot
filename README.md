# currency-telegram-bot

## 🚀 Deployment on Railway

### Automatic Deployment (Recommended)
1. Fork this repository
2. Go to [Railway](https://railway.app) and create account
3. Click "New Project" → "Deploy from GitHub repo"
4. Connect your forked repository
5. Add environment variables:
   - `BOT_TOKEN` - your Telegram bot token
   - `DB_URL` - PostgreSQL connection string (Railway provides this automatically)
   - `LOG_LEVEL` - `info` (recommended for production)
   - `CACHE_TTL_MINUTES` - `5`

### Manual Deployment
1. Install Railway CLI: `npm i -g @railway/cli`
2. Login: `railway login`
3. Link project: `railway link`
4. Set environment variables: `railway variables set BOT_TOKEN=your_token`
5. Deploy: `railway up`