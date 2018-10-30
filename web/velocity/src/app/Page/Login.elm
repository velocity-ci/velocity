module Page.Login exposing (ExternalMsg(..), Model, Msg, initialModel, update, view)

{-| The login page.
-}

import Component.Form exposing (ifBelowLength, validClasses)
import Context exposing (Context)
import Data.Session as Session exposing (Session)
import Data.User as User exposing (User)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Request.Errors
import Request.User exposing (storeSession)
import Route exposing (Route)
import Task
import Util exposing ((=>))
import Validate exposing (..)
import Views.Form as Form


-- MODEL --


type Field
    = Form
    | Username
    | Password


type alias FormField =
    { value : String
    , dirty : Bool
    , field : Field
    }


type alias Model =
    { errors : List Error
    , username : FormField
    , password : FormField
    , submitting : Bool
    , globalError : Maybe String
    }


newField : Field -> FormField
newField field =
    FormField "" False field


initialModel : Model
initialModel =
    let
        initial =
            { errors = []
            , username = newField Username
            , password = newField Password
            , submitting = False
            , globalError = Nothing
            }
    in
        { initial | errors = validate initial }



-- VIEW --


view : Session msg -> Model -> Html Msg
view session model =
    div [ class "d-flex justify-content-center", (\( a, b ) -> style a b) ("height" => "100vh") ]
        [ div [ class "card col-8 align-self-center" ]
            [ div [ class "card-body" ]
                [ viewGlobalError model.globalError
                , viewForm model
                ]
            ]
        ]


viewGlobalError : Maybe String -> Html Msg
viewGlobalError maybeError =
    case maybeError of
        Just error ->
            div [ class "alert alert-danger", attribute "role" "alert" ]
                [ text error ]

        Nothing ->
            text ""


viewForm : Model -> Html Msg
viewForm model =
    let
        inputClassList =
            validClasses <| model.errors
    in
        Html.form [ attribute "novalidate" "", onSubmit SubmitForm ]
            [ Form.input
                { name = "username"
                , label = "Username"
                , help = Nothing
                , errors = []
                }
                [ classList <| inputClassList model.username
                , attribute "required" ""
                , onInput SetUsername
                , value model.username.value
                ]
                []
            , Form.password
                { name = "password"
                , label = "Password"
                , help = Nothing
                , errors = []
                }
                [ classList <| inputClassList model.password
                , attribute "required" ""
                , onInput SetPassword
                , value model.password.value
                ]
                []
            , button
                [ class "btn btn-primary"
                , type_ "submit"
                , disabled ((not <| List.isEmpty model.errors) && (not <| model.submitting))
                ]
                [ text "Submit" ]
            , Util.viewIf model.submitting Form.viewSpinner
            ]



-- UPDATE --


type Msg
    = SubmitForm
    | SetUsername String
    | SetPassword String
    | LoginCompleted (Result Request.Errors.HttpError User)


type ExternalMsg
    = NoOp
    | SetUser User


updateInput : Field -> String -> FormField
updateInput field value =
    FormField value True field


update : Context -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update context msg model =
    case msg of
        SubmitForm ->
            case validate model of
                [] ->
                    let
                        submitValues =
                            { username = model.username.value
                            , password = model.password.value
                            }
                    in
                        { model
                            | errors = []
                            , submitting = True
                            , globalError = Nothing
                        }
                            => Task.attempt LoginCompleted (Request.User.login context submitValues)
                            => NoOp

                errors ->
                    { model | errors = errors }
                        => Cmd.none
                        => NoOp

        SetUsername username ->
            let
                newModel =
                    { model | username = username |> updateInput Username }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none
                    => NoOp

        SetPassword password ->
            let
                newModel =
                    { model | password = password |> updateInput Password }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none
                    => NoOp

        LoginCompleted (Err error) ->
            let
                newModel =
                    { model | password = newField Password }
            in
                { newModel
                    | errors = validate newModel
                    , submitting = False
                    , globalError = Just "Invalid username or password. Please try again..."
                }
                    => Cmd.none
                    => NoOp

        LoginCompleted (Ok user) ->
            { model | submitting = False }
                => Cmd.batch [ storeSession user, Route.modifyUrl Route.Home ]
                => SetUser user



-- VALIDATION --


type alias Error =
    ( Field, String )


validate : Model -> List Error
validate =
    Validate.all
        [ (.username >> .value) >> ifBlank (Username => "username can't be blank.")
        , (.password >> .value) >> ifBlank (Password => "password can't be blank.")
        , (.username >> .value) >> ifBelowLength 3 (Username => "username must be over 2 characters.")
        , (.password >> .value) >> ifBelowLength 3 (Password => "password must be over 2 characters.")
        ]
