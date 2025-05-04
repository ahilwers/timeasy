# Stage 1: Build
FROM node as build
WORKDIR /app
COPY package*.json ./
RUN npm install
RUN npm install -g @angular/cli
COPY . .

# Stage 2: Run
RUN ng build --configuration=production
FROM nginx:latest
COPY --from=build app/dist/timeasy-webclient /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
