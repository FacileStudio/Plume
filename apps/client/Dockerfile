FROM oven/bun:1 AS build

WORKDIR /app

COPY package.json bun.lock* ./

RUN bun install --ignore-scripts

COPY . .

RUN bunx svelte-kit sync && bun run build

FROM oven/bun:1-slim

WORKDIR /app

COPY --from=build /app/build ./build
COPY --from=build /app/package.json ./
COPY --from=build /app/node_modules ./node_modules

ENV PORT=3000
ENV API_URL=http://plume-api:4000

EXPOSE 3000

CMD ["bun", "build/index.js"]
