locals {
    topology = data.terraform_remote_state.prod.outputs.topology
}

data "aws_apigatewayv2_apis" "kaixo" {
  name = "kaixo"
}

data "aws_apigatewayv2_api" "kaixo" {
  api_id = tolist(data.aws_apigatewayv2_apis.kaixo.ids)[0]
}

data "aws_lambda_function" "binnit" {
  function_name = "binnit-main-src"
}

data "aws_iam_policy_document" "invoke" {
  statement {
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction"
    ]
    resources = [
      data.aws_lambda_function.binnit.arn
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "execute-api:Invoke"
    ]
    resources = [
      "${data.aws_apigatewayv2_api.kaixo.execution_arn}/*/binnit/main/src/*"
    ]
  }
}

module "invoke" {
  source = "github.com/bkeane/stage/stage?ref=v0.1.1"
  stage                    = "invoke"
  topology                 = local.topology
  policy_document          = data.aws_iam_policy_document.invoke
}