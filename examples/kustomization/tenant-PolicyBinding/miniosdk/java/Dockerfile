
FROM openjdk:11

RUN mkdir /app
WORKDIR /app
COPY target/operator-sts-0.1.0.jar /app/app.jar

CMD ["java", "-jar", "./app/app.jar"]
