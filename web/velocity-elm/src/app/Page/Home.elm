module Page.Home exposing (view, update, Model, Msg, initialModel)

{-| The homepage. You can get here via either the / or /#/ routes.
-}

import Html exposing (..)
import Html.Attributes exposing (class, href, id, placeholder, attribute, classList)
import Data.Session as Session exposing (Session)
import Util exposing ((=>), onClickStopPropagation)


-- MODEL --


initialModel : Model
initialModel =
    {}


type alias Model =
    {}


view : Session -> Model -> Html Msg
view session model =
    div [ class "row" ]
        [ div [ class "col-12 col-md-6" ]
            [ div [ class "card" ]
                [ h4 [ class "card-header" ]
                    [ text "Last builds" ]
                , ul [ class "list-group" ]
                    [ li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small []
                                [ text "3 days ago" ]
                            ]
                        , p [ class "mb-1" ]
                            [ text "Donec id elit non mi porta gravida at eget metus. Maecenas sed diam eget risus varius blandit." ]
                        , small []
                            [ text "Donec id elit non mi porta." ]
                        ]
                    , li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small [ class "text-muted" ]
                                [ text "3 days ago" ]
                            ]
                        ]
                    , li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small [ class "text-muted" ]
                                [ text "3 days ago" ]
                            ]
                        ]
                    ]
                ]
            ]
        , div [ class "col-12 col-md-6" ]
            [ div [ class "card" ]
                [ h4 [ class "card-header" ]
                    [ text "Projects" ]
                , ul [ class "list-group" ]
                    [ li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small []
                                [ text "3 days ago" ]
                            ]
                        , p [ class "mb-1" ]
                            [ text "Donec id elit non mi porta gravida at eget metus. Maecenas sed diam eget risus varius blandit." ]
                        , small []
                            [ text "Donec id elit non mi porta." ]
                        ]
                    , li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small [ class "text-muted" ]
                                [ text "3 days ago" ]
                            ]
                        , p [ class "mb-1" ]
                            [ text "Donec id elit non mi porta gravida at eget metus. Maecenas sed diam eget risus varius blandit." ]
                        , small [ class "text-muted" ]
                            [ text "Donec id elit non mi porta." ]
                        ]
                    , li [ class "list-group-item list-group-item-action flex-column align-items-start" ]
                        [ div [ class "d-flex w-100 justify-content-between" ]
                            [ h5 [ class "mb-1" ]
                                [ text "List group item heading" ]
                            , small [ class "text-muted" ]
                                [ text "3 days ago" ]
                            ]
                        , p [ class "mb-1" ]
                            [ text "Donec id elit non mi porta gravida at eget metus. Maecenas sed diam eget risus varius blandit." ]
                        , small [ class "text-muted" ]
                            [ text "Donec id elit non mi porta." ]
                        ]
                    ]
                ]
            ]
        ]



-- UPDATE --


type Msg
    = NoOp


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    model => Cmd.none
