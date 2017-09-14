module Page.Login exposing (view, update, Model, Msg, initialModel, ExternalMsg(..))

{-| The login page.
-}

import Route exposing (Route)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Views.Form as Form
import Json.Decode as Decode exposing (field, decodeString, string, Decoder)
import Json.Decode.Pipeline as Pipeline exposing (optional, decode)
import Validate exposing (..)
import Data.Session as Session exposing (Session)
import Http
import Request.User exposing (storeSession)
import Util exposing ((=>))
import Data.User as User exposing (User)


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
            }
    in
        { initial | errors = validate initial }



-- VIEW --


view : Session -> Model -> Html Msg
view session model =
    div [ class "row justify-content-md-center" ]
        [ div [ class "col col-md-6" ]
            [ div [ class "card" ]
                [ div [ class "card-body" ]
                    [ viewForm model ]
                ]
            ]
        ]


isInvalid : List Error -> FormField -> Bool
isInvalid errors formField =
    formField.dirty && List.length (getFieldErrors formField errors) > 0


isValid : List Error -> FormField -> Bool
isValid errors formField =
    formField.dirty && List.length (getFieldErrors formField errors) == 0


validClasses : List Error -> FormField -> List ( String, Bool )
validClasses errors field =
    [ ( "is-invalid", isInvalid errors field )
    , ( "is-valid", isValid errors field )
    ]


viewForm : Model -> Html Msg
viewForm model =
    Html.form [ attribute "novalidate" "", onSubmit SubmitForm ]
        [ div [ class "form-group" ]
            [ label [ for "username" ] [ text "Username" ]
            , input
                [ class "form-control"
                , classList (validClasses model.errors model.username)
                , id "username"
                , placeholder "Username"
                , attribute "required" ""
                , type_ "text"
                , onInput SetUsername
                , value model.username.value
                ]
                []
            ]
        , div [ class "form-group" ]
            [ label [ for "password" ] [ text "Password" ]
            , input
                [ class "form-control"
                , classList (validClasses model.errors model.password)
                , id "password"
                , placeholder "Password"
                , attribute "required" ""
                , type_ "password"
                , onInput SetPassword
                , value model.password.value
                ]
                []
            ]
        , button
            [ class "btn btn-primary"
            , type_ "submit"
            , disabled (not <| List.isEmpty model.errors)
            ]
            [ text "Submit" ]
        , Util.viewIf model.submitting viewFormLoadingSpinner
        ]


viewFormLoadingSpinner : Html Msg
viewFormLoadingSpinner =
    span []
        [ i [ class "fa fa-circle-o-notch fa-spin fa-fw" ] []
        , span [ class "sr-only" ] [ text "Loading..." ]
        ]



-- UPDATE --


type Msg
    = SubmitForm
    | SetUsername String
    | SetPassword String
    | LoginCompleted (Result Http.Error User)


type ExternalMsg
    = NoOp
    | SetUser User


updateInput : Field -> String -> FormField
updateInput field value =
    { value = value, field = field, dirty = True }


inputValue : FormField -> String
inputValue input =
    input.value


getFieldErrors : FormField -> List Error -> List Error
getFieldErrors formField errors =
    let
        isFieldError error =
            let
                ( field, _ ) =
                    error
            in
                formField.field == field
    in
        List.filter isFieldError errors


update : Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update msg model =
    case msg of
        SubmitForm ->
            case validate model of
                [] ->
                    let
                        submitValues =
                            { username = inputValue model.username, password = inputValue model.password }
                    in
                        { model | errors = [], submitting = True }
                            => Http.send LoginCompleted (Request.User.login submitValues)
                            => NoOp

                errors ->
                    { model | errors = errors }
                        => Cmd.none
                        => NoOp

        SetUsername username ->
            let
                newModel =
                    { model | username = username |> (updateInput Username) }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none
                    => NoOp

        SetPassword password ->
            let
                newModel =
                    { model | password = password |> (updateInput Password) }
            in
                { newModel | errors = validate newModel }
                    => Cmd.none
                    => NoOp

        LoginCompleted (Err error) ->
            let
                errorMessages =
                    case error of
                        Http.BadStatus response ->
                            response.body
                                |> decodeString (field "errors" errorsDecoder)
                                |> Result.withDefault []

                        _ ->
                            [ "unable to process registration" ]
            in
                { model
                    | errors = List.map (\errorMessage -> Form => errorMessage) errorMessages
                    , submitting = False
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


ifBelowLength : Int -> error -> Validator error String
ifBelowLength length =
    ifInvalid (\s -> String.length s < length)


validate : Model -> List Error
validate =
    Validate.all
        [ (.username >> .value) >> ifBlank (Username => "username can't be blank.")
        , (.password >> .value) >> ifBlank (Password => "password can't be blank.")
        , (.username >> .value) >> (ifBelowLength 3) (Username => "username must be over 3 characters.")
        , (.password >> .value) >> (ifBelowLength 8) (Password => "password must be over 8 characters.")
        ]


errorsDecoder : Decoder (List String)
errorsDecoder =
    decode (\username password -> List.concat [ username, password ])
        |> optionalError "username"
        |> optionalError "password"


optionalError : String -> Decoder (List String -> a) -> Decoder a
optionalError fieldName =
    let
        errorToString errorMessage =
            String.join " " [ fieldName, errorMessage ]
    in
        optional fieldName (Decode.list (Decode.map errorToString string)) []
