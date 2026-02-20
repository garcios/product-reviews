FROM node:24-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY index.js .
CMD ["node", "index.js"]
