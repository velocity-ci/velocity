module Page.Project.Commit.Build exposing (..)

import Data.Build as Build exposing (Build)
import Data.Project as Project exposing (Project)
import Data.Commit as Commit exposing (Commit)


--import Data.LogOutput as LogOutput exposing (LogOutput)

import Html exposing (..)
import Util exposing ((=>))
import Socket.Channel as Channel exposing (Channel)
import Json.Encode as Encode
import Json.Decode as Decode


-- SUBSCRIPTIONS --
--channel : Build -> Channel Msg
--channel build =
--    let
--        buildChannelPath =
--            [ "project"
--            , Project.idToString build.project
--            , "commits"
--            , Commit.hashToString build.commit
--            , "builds"
--            , Build.idToString build.id
--            ]
--    in
--        buildChannelPath
--            |> String.join "/"
--            |> Channel.init
--            |> Channel.onJoin SocketUpdate
-- MODEL --


type alias Model =
    { build : Build
    }


initialModel : Build -> Model
initialModel build =
    { build =
        build
        --    , output = []
    }



-- VIEW --


view : Model -> Html Msg
view model =
    div []
        [ text "BUILD VIEW"
        ]



-- UPDATE --


type Msg
    = NoOp
    | SocketUpdate Encode.Value



--responseDecoder : Decode.Decoder (List LogOutput)
--responseDecoder =
--    Decode.field "log" LogOutput.decoder
--        |> Decode.list


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        --        SocketUpdate json ->
        --            { model | output = Decode.decodeValue responseDecoder json |> Result.withDefault [] }
        --                => Cmd.none
        _ ->
            model => Cmd.none
