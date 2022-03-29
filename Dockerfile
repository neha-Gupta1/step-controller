FROM alpine
COPY step-controller /usr/local/bin
RUN echo $PATH 
ENTRYPOINT ["step-controller"]