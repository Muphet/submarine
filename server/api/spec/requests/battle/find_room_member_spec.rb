require 'rails_helper'

RSpec.describe "Battle::FindRoomMember", type: :request do
  describe "POST /battle/find_room_member" do
    let(:room_member) { create(:room_member) }
    let(:request_params) { { room_key: room_member.room_key } }

    before do
      post battle_find_room_member_path, params: request_params
    end

    context 'with a valid request' do
      it 'should work' do
        expect(response).to have_http_status(200)
      end
      it 'should return the requested room' do
        expect(parsed_response.room_member.id).to eq room_member.user.id
      end
    end

    context 'with a no-existing room_key' do
      let(:request_params) { { room_key: "" } }

      it 'should return nil as the requested room' do
        expect(parsed_response.room_member).to eq nil
      end
    end
  end
end
