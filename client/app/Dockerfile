# Pull official base image
FROM node:16-alpine

# Set working directory
WORKDIR /app

# Install app dependencies
COPY package.json ./
COPY package-lock.json ./

RUN npm install --silent

# Add app
COPY . ./

# Start app
CMD ["npm", "start"]
