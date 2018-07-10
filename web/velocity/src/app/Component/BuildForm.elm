module Component.BuildForm
    exposing
        ( Context
        , Config
        , init
        , Field(..)
        , InputFormField
        , ChoiceFormField
        , updateInput
        , updateSelect
        , submitParams
        , view
        , viewSubmitButton
        , firstId
        )

-- EXTERNAL --

import Validate exposing (..)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onSubmit, on, onInput, onClick)
import Html.Events.Extra exposing (targetSelectedIndex)
import Json.Decode as Decode
import Bootstrap.Button as Button


-- INTERNAL --

import Util exposing ((=>))
import Data.Task as ProjectTask exposing (Step(..), Parameter(..))
import Views.Form as Form
import Component.Form exposing (validClasses)


-- MODEL --


type alias InputFormField =
    { value : String
    , dirty : Bool
    , field : String
    }


type alias ChoiceFormField =
    { value : Maybe String
    , dirty : Bool
    , field : String
    , options : List String
    }


type Field
    = Input InputFormField
    | Choice ChoiceFormField


type alias Context =
    { fields : List Field
    , errors : List Error
    }


type alias Config msg =
    { submitMsg : msg
    , onChangeMsg : ChoiceFormField -> Maybe Int -> msg
    , onInputMsg : InputFormField -> String -> msg
    }


init : ProjectTask.Task -> Context
init task =
    let
        fields =
            List.filterMap newField task.parameters

        errors =
            List.concatMap validator fields
    in
        { fields = fields
        , errors = errors
        }


newField : Parameter -> Maybe Field
newField parameter =
    case parameter of
        StringParam param ->
            let
                value =
                    Maybe.withDefault "" param.default

                dirty =
                    String.length value > 0
            in
                InputFormField value dirty param.name
                    |> Input
                    |> Just

        ChoiceParam param ->
            let
                options =
                    param.default
                        :: (List.map Just param.options)
                        |> List.filterMap identity

                value =
                    case param.default of
                        Nothing ->
                            List.head options

                        default ->
                            default
            in
                ChoiceFormField value True param.name options
                    |> Choice
                    |> Just

        DerivedParam _ ->
            Nothing



-- UPDATE --


updateInput : InputFormField -> String -> Context -> Context
updateInput field value context =
    let
        updateField fieldType =
            case fieldType of
                Input f ->
                    if f == field then
                        Input
                            { field
                                | value = value
                                , dirty = True
                            }
                    else
                        fieldType

                _ ->
                    fieldType

        fields =
            List.map updateField context.fields

        errors =
            List.concatMap validator fields
    in
        { context
            | fields = fields
            , errors = errors
        }


updateSelect : ChoiceFormField -> Maybe Int -> Context -> Context
updateSelect field maybeIndex context =
    let
        updateField fieldType =
            case ( fieldType, maybeIndex ) of
                ( Choice f, Just index ) ->
                    if f == field then
                        let
                            value =
                                f.options
                                    |> List.indexedMap (,)
                                    |> List.filter (\( i, _ ) -> i == index)
                                    |> List.head
                                    |> Maybe.map Tuple.second
                        in
                            Choice
                                { field
                                    | value = value
                                    , dirty = True
                                }
                    else
                        fieldType

                _ ->
                    fieldType

        fields =
            List.map updateField context.fields

        errors =
            List.concatMap validator fields
    in
        { context
            | fields = fields
            , errors = errors
        }


submitParams : Context -> List ( String, String )
submitParams { fields } =
    let
        stringParam { value, field } =
            field => value

        mapFieldToParam field =
            case field of
                Input input ->
                    Just (stringParam input)

                Choice choice ->
                    choice.value
                        |> Maybe.map (\value -> stringParam { value = value, field = choice.field })
    in
        List.filterMap mapFieldToParam fields



-- VIEW --


view : Config msg -> Context -> List (Html msg)
view config { fields, errors } =
    [ Html.form [ class "mt-3", attribute "novalidate" "", onSubmit config.submitMsg ] <|
        List.map (viewField config errors) fields
    ]


viewField :
    { a
        | onChangeMsg : ChoiceFormField -> Maybe Int -> msg
        , onInputMsg : InputFormField -> String -> msg
    }
    -> List ( String, error )
    -> Field
    -> Html msg
viewField { onChangeMsg, onInputMsg } errors f =
    case f of
        Choice field ->
            viewChoiceField onChangeMsg errors field

        Input field ->
            viewInputField onInputMsg errors field


viewChoiceField : (ChoiceFormField -> Maybe Int -> msg) -> List ( String, error ) -> ChoiceFormField -> Html msg
viewChoiceField onChangeMsg errors field =
    let
        value =
            Maybe.withDefault "" field.value

        option o =
            Html.option
                [ selected (o == value) ]
                [ text o ]
    in
        Form.select
            { name = field.field
            , label = field.field
            , help = Nothing
            , errors = []
            }
            [ attribute "required" ""
            , classList (validClasses errors field)
            , on "change" <| Decode.map (onChangeMsg field) targetSelectedIndex
            ]
            (List.map option field.options)


viewInputField : (InputFormField -> String -> msg) -> List ( String, error ) -> InputFormField -> Html msg
viewInputField onInputMsg errors field =
    Form.input
        { name = field.field
        , label = field.field
        , help = Nothing
        , errors = []
        }
        [ attribute "required" ""
        , value field.value
        , onInput (onInputMsg field)
        , classList (validClasses errors field)
        ]
        []


viewSubmitButton : Config msg -> Context -> Html msg
viewSubmitButton { submitMsg } { errors } =
    let
        hasErrors =
            not <| List.isEmpty errors
    in
        Button.button
            [ Button.outlinePrimary
            , Button.attrs
                [ onClick submitMsg
                , disabled hasErrors
                ]
            ]
            [ text "Start" ]



-- HELPERS --


firstId : Context -> Maybe String
firstId context =
    case List.head context.fields of
        Just (Input { field }) ->
            Just field

        Just (Choice { field }) ->
            Just field

        Nothing ->
            Nothing



-- VALIDATION --


type alias Error =
    ( String, String )


validator : Validator Error Field
validator =
    [ \f ->
        let
            notBlank { field, value } =
                ifBlank (field => "Field cannot be blank") value
        in
            case f of
                Input fieldType ->
                    notBlank fieldType

                Choice fieldType ->
                    (\{ field, value } ->
                        value
                            |> Maybe.withDefault ""
                            |> ifBlank (field => "Field cannot be blank")
                    )
                        fieldType
    ]
        |> Validate.all
