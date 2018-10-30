module Component.UserForm
    exposing
        ( Config
        , Context
        , errorsDecoder
        , init
        , password
        , passwordConfirm
        , serverErrorToFormError
        , submitValues
        , update
        , username
        , view
        , viewSubmitButton
        )

-- EXTERNAL --
-- INTERNAL --

import Bootstrap.Button as Button
import Component.Form as BaseForm
    exposing
        ( FormField
        , ifBelowLength
        , newField
        , optionalError
        , resetServerErrorsForField
        , updateInput
        )
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput, onSubmit)
import Json.Decode as Decode exposing (Decoder, decodeString, field, string)
import Json.Decode.Pipeline as Pipeline exposing (decode, optional)
import Util exposing ((=>))
import Validate exposing (..)
import Views.Form as Form


type Field
    = Form
    | Username
    | Password
    | PasswordConfirm


type alias UserForm =
    { username : FormField Field
    , password : FormField Field
    , passwordConfirm : FormField Field
    }


type alias Context =
    BaseForm.Context Field UserForm


type alias Error =
    BaseForm.Error Field


type alias Config msg =
    { setUsernameMsg : String -> msg
    , setPasswordMsg : String -> msg
    , setPasswordConfirmMsg : String -> msg
    , submitMsg : msg
    }


username : Field
username =
    Username


password : Field
password =
    Password


passwordConfirm : Field
passwordConfirm =
    PasswordConfirm


initialForm : UserForm
initialForm =
    { username = newField Username
    , password = newField Password
    , passwordConfirm = newField PasswordConfirm
    }


init : Context
init =
    { form = initialForm
    , errors = []
    , serverErrors = []
    , submitting = False
    }



-- UPDATE HELPERS --


updateForm : UserForm -> Field -> String -> UserForm
updateForm form field value =
    let
        updatedField =
            updateInput field value
    in
        case field of
            Username ->
                { form | username = updatedField }

            Password ->
                { form | password = updatedField }

            PasswordConfirm ->
                { form | passwordConfirm = updatedField }

            Form ->
                form


update : Context -> Field -> String -> Context
update context field value =
    let
        updatedForm =
            updateForm context.form field value
    in
        { context
            | errors = validate updatedForm
            , form = updatedForm
            , serverErrors = resetServerErrorsForField context field
        }


submitValues : Context -> { username : String, password : String }
submitValues { form } =
    { username = form.username.value
    , password = form.password.value
    }


serverErrorToFormError : ( String, String ) -> Error
serverErrorToFormError ( fieldNameString, errorString ) =
    let
        field =
            case fieldNameString of
                "username" ->
                    Username

                "password" ->
                    Password

                _ ->
                    Form
    in
        ( field, errorString )



-- VIEW --


view : Config msg -> Context -> Html msg
view config context =
    div []
        (List.map viewGlobalError (BaseForm.globalErrors Form (BaseForm.allErrors context))
            ++ [ Html.form [ attribute "novalidate" "", onSubmit config.submitMsg ]
                    [ viewUsername context config.setUsernameMsg
                    , viewPassword context config.setPasswordMsg
                    , viewPasswordConfirm context config.setPasswordConfirmMsg
                    ]
               ]
        )


viewGlobalError : Error -> Html msg
viewGlobalError error =
    div [ class "alert alert-danger" ] [ text (Tuple.second error) ]


viewUsername : Context -> (String -> msg) -> Html msg
viewUsername ({ form } as context) msg =
    Form.input
        { name = "username"
        , label = "Username"
        , help = Nothing
        , errors = errors form.username context
        }
        [ attribute "required" ""
        , value form.username.value
        , onInput msg
        , classList <| fieldClassList context form.username
        , disabled context.submitting
        ]
        []


viewPassword : Context -> (String -> msg) -> Html msg
viewPassword ({ form } as context) msg =
    Form.password
        { name = "password"
        , label = "Password"
        , help = Nothing
        , errors = errors form.password context
        }
        [ attribute "required" ""
        , value form.password.value
        , onInput msg
        , classList <| fieldClassList context form.password
        , disabled context.submitting
        ]
        []


viewPasswordConfirm : Context -> (String -> msg) -> Html msg
viewPasswordConfirm ({ form } as context) msg =
    Form.password
        { name = "password_confirm"
        , label = "Confirm password"
        , help = Nothing
        , errors = errors form.passwordConfirm context
        }
        [ attribute "required" ""
        , value form.passwordConfirm.value
        , onInput msg
        , classList <| fieldClassList context form.passwordConfirm
        , disabled context.submitting
        ]
        []


viewSubmitButton : Config msg -> Context -> Html msg
viewSubmitButton { submitMsg } ({ form } as context) =
    let
        hasErrors =
            not <| List.isEmpty context.errors

        submitting =
            context.submitting

        untouched =
            not form.username.dirty && not form.password.dirty && not form.passwordConfirm.dirty
    in
        Button.button
            [ Button.outlinePrimary
            , Button.attrs
                [ onClick submitMsg
                , disabled (hasErrors || submitting || untouched)
                ]
            ]
            [ text "Create" ]


fieldClassList :
    BaseForm.Context field msg
    -> { formField | dirty : Bool, field : field }
    -> List ( String, Bool )
fieldClassList context =
    BaseForm.validClasses (BaseForm.allErrors context)


errors :
    { a | dirty : Bool, field : field }
    -> BaseForm.Context field msg
    -> List ( field, String )
errors field context =
    if field.dirty then
        BaseForm.getFieldErrors (BaseForm.allErrors context) field
    else
        []


passwordValidator : ( { a | value : String }, { b | value : String } ) -> Bool
passwordValidator ( a, b ) =
    a.value /= b.value



-- VALIDATION --


validate : Validator ( Field, String ) UserForm
validate =
    Validate.all
        [ (.username >> .value) >> ifBelowLength 3 (Username => "Username must be over 2 characters.")
        , (.password >> .value) >> ifBelowLength 3 (Password => "Password must be over 2 characters.")
        , (\f -> ( f.password, f.passwordConfirm )) >> ifInvalid passwordValidator (PasswordConfirm => "Passwords do not match")
        ]


errorsDecoder : Decoder (List ( String, String ))
errorsDecoder =
    decode (\username password -> List.concat [ username, password ])
        |> optionalError "username"
        |> optionalError "password"
