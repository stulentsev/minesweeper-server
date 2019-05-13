#! /usr/bin/env ruby
require 'bundler'
Bundler.require

class ApiClient

  def initialize(host)
    @host = host
    @connection = Excon.new(host)
  end

  def new_game
    response = post("/newgame")
    @game_id = response['game_id']
    puts "Started game: #{game_id}"

  end

  def move(x, y)
    resp = post("/move", { game_id: game_id, x: x, y: y })
    if resp['pretty_board_state']
      puts resp['pretty_board_state']
    else
      puts resp
    end
    puts resp
  end

  private

  attr_reader :host, :connection, :game_id

  def post(path, payload = {})
    request(path, 'POST', payload)
  end

  def request(path, method, payload)
    response = connection.request(
      path: path,
      method: method,
      body: payload.to_json,
    )

    JSON.parse(response.body)
  end

end

@client = ApiClient.new("http://localhost:3000")

def new_game
  @client.new_game
end


def move(x, y)
  @client.move(x, y)
end

Pry.start
