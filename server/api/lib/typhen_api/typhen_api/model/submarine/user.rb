# This file was generated by typhen-api

module TyphenApi::Model::Submarine
  class User
    include Virtus.model(:strict => true)

    attribute :id, Integer, :required => true
    attribute :name, String, :required => true
  end
end
