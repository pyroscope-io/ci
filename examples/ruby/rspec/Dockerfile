FROM ruby:3.2.1

WORKDIR /usr/src/app

COPY Gemfile ./
COPY Gemfile.lock ./
COPY fib.rb ./
COPY spec spec/

RUN bundle install --binstubs

CMD ["ruby", "main.rb"]

