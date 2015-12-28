class LoginController < ApplicationController
  include TyphenApi::Controller::Submarine::Login
  include TyphenApiRespondable

  def service
    if logged_in_user.blank?
      raise ApplicationError::LoginFailed.new('The user name or password is incorrect')
    end

    render_response(
      user: logged_in_user.to_api_type,
      joined_room: joined_room.try(:to_api_type))
  end

  def logged_in_user
    @logged_in_user ||= login(params.name, params.password)
  end

  def joined_room
    @joinned_room ||= logged_in_user.room
  end
end
