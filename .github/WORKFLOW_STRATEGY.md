# GitHub Workflow Strategy

This document outlines the branch and workflow strategy for this project.

## Branch Strategy

### Main Branch (`main`)
- **Protected branch** - no direct pushes allowed
- **Integration branch** - all features merge here via PRs
- **Release source** - all releases originate from main
- **Requires PR approval** and all checks to pass

### Feature Branches
- **Created from `main`** for each new task/feature
- **Any branch name** can be used (feature/, fix/, chore/, etc.)
- **Workflows run automatically** on push for early validation
- **Merged to `main`** via Pull Request after all checks pass

### Develop Branch (`develop`) - Optional
- Used for integration testing if needed
- Does not trigger releases
- Can be used for staging/testing purposes

## Workflow Strategy

### 1. Development Flow
```
main branch (protected)
    ↓
feature/fix branch (from main)
    ↓
Push to feature branch → Workflows run ✅
    ↓
All checks pass → Create Pull Request to main
    ↓
PR workflows pass ✅
    ↓
PR merged to main
    ↓
Release process triggered
```

### 2. Branch Validation Workflows

**On ANY branch push:**
- ✅ **CI Pipeline** (lint, test, build, security) - validates code quality
- ✅ **Security Scanning** - early security validation
- ✅ **Multi-Arch Build** - validates multi-platform compatibility (build only)
- ❌ **Release Process** - only on main

**On PR to main:**
- ✅ **All CI checks** must pass
- ✅ **PR Validation** (title, labels, description)
- ✅ **Security Scanning** 
- ✅ **Code Review** required

### 3. Release Process

Only triggered when PR is merged to `main`:
```
PR merged to main
    ↓
CI Workflow ✅ (on main)
    ↓
Multi-Arch Build ✅
    ↓
Semantic Release ✅
    ↓
Release Published ✅
    ↓
Helm Chart & Containers Published
```

## Workflow Triggers

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to any branch, PR to main | Validate code quality early and on PR |
| `pr-validation.yml` | PR to main | Validate PR standards |
| `security.yml` | Push to any branch, PR to main, scheduled | Security scanning everywhere |
| `multi-arch-build.yml` | Push to any branch, CI success on main | Multi-platform validation (build on all branches, push on main) |
| `semantic-release.yml` | Multi-arch build success | Create releases |
| `release.yml` | Release published | Publish containers & Helm |

## Developer Workflow

### 1. Start New Work
```bash
git checkout main
git pull origin main
git checkout -b feature/your-feature-name
```

### 2. Develop & Validate
```bash
# Make changes
git add .
git commit -m "feat: add new feature"
git push origin feature/your-feature-name

# ✅ Workflows automatically run and validate your changes
# ✅ CI: lint, test, build, security
# ✅ Security: vulnerability scanning
# ✅ Multi-Arch: build for linux/amd64 and linux/arm64 (validates compatibility)
# Check GitHub Actions tab for results
```

### 3. Test Multi-Platform Images (Optional)
```bash
# If you need to test the actual images, trigger a build with push enabled
# Go to GitHub Actions → Multi-Arch Build → Run workflow
# Set "Push images to registry" to true
# This will push images tagged as: ghcr.io/your-repo:feature-branch-name

# Pull and test your images locally
docker pull ghcr.io/your-org/your-repo:feature-your-feature-name
docker run --rm ghcr.io/your-org/your-repo:feature-your-feature-name --help
```

### 4. Create PR (only when all checks pass)
```bash
# Only create PR if your branch workflows are green ✅
# Go to GitHub and create PR to main
# All PR workflows must pass before merge
```

### 5. After Merge
```bash
# Automatic release process triggered
# Clean up your feature branch
git checkout main
git pull origin main
git branch -d feature/your-feature-name
```

## Benefits of This Strategy

✅ **Early Validation** - Catch issues before creating PRs
✅ **Developer Confidence** - Know your changes work before PR
✅ **Faster PR Reviews** - Reviewers see pre-validated code  
✅ **Reduced CI Load** - Failed builds caught early, not on main
✅ **Better Quality** - Multiple validation layers
✅ **Multi-Platform Testing** - Validate ARM64 and AMD64 compatibility early
✅ **Container Validation** - Test actual runtime behavior before release

## Branch Protection Settings

Configure the following branch protection rules for `main`:

```yaml
# GitHub Settings > Branches > Add rule
Branch name pattern: main
Settings:
  ✅ Require a pull request before merging
  ✅ Require approvals (1)
  ✅ Dismiss stale PR approvals when new commits are pushed
  ✅ Require review from code owners
  ✅ Require status checks to pass before merging
  ✅ Require branches to be up to date before merging
  ✅ Require conversation resolution before merging
  ✅ Restrict pushes that create files larger than 100MB
  ✅ Allow force pushes: Never
  ✅ Allow deletions: Never
```

## Required Status Checks

Add these as required status checks for `main` branch:
- `CI Success`
- `PR Validation Summary`
- `Security Summary`
- `Code quality / CodeQL Analysis` 

## Multi-Arch Build Strategy

### **Feature Branch Builds**
- ✅ **Build** multi-platform images (linux/amd64, linux/arm64)
- ✅ **Test** platform compatibility
- ✅ **Validate** container functionality  
- ❌ **No automatic push** (saves registry space and costs)
- 🔧 **Manual push** available via workflow dispatch

### **Main Branch Builds**
- ✅ **Build** multi-platform images
- ✅ **Push** development images automatically
- ✅ **Test** across platforms
- ✅ **Trigger** semantic release process

### **Image Tagging Strategy**
```bash
# Feature branches (manual push only)
ghcr.io/your-org/your-repo:feature-branch-name
ghcr.io/your-org/your-repo:fix-bug-123

# Main branch (automatic)
ghcr.io/your-org/your-repo:main
ghcr.io/your-org/your-repo:main-abc123

# Releases (via Release workflow)
ghcr.io/your-org/your-repo:v1.2.3
ghcr.io/your-org/your-repo:latest
``` 