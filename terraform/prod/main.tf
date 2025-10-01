data "aws_apigatewayv2_apis" "kaixo" {
  name = "kaixo"
}

data "aws_apigatewayv2_api" "kaixo" {
  api_id = tolist(data.aws_apigatewayv2_apis.kaixo.ids)[0]
}

data "aws_lambda_function" "binnit" {
  function_name = "binnit-main-src"
}

resource "aws_ecr_repository" "qurl" {
  name = "bkeane/qurl/cmd"
}

module "topology" {
  source = "github.com/bkeane/stage/topology?ref=v0.1.1"
  origin = "https://github.com/bkeane/qurl.git"
  
  accounts = {
    "prod" = "677771948337"
    "dev" = "831926600600"
  }

  ecr_repositories = [
    aws_ecr_repository.qurl
  ]

  stages = [
    "invoke"
  ]
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
  topology                 = module.topology
  policy_document          = data.aws_iam_policy_document.invoke
}

resource "local_file" "action" {
  content = module.topology.action
  filename = "../../../.github/actions/stages/action.yaml"
}

output "topology" {
  value = module.topology
}
