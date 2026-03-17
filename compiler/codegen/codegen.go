package codegen

import (
	"fmt"
	"strings"

	"github.com/annuaicoder/yoft/compiler/ast"
)

type Generator struct {
	out        strings.Builder
	indent     int
	varCounter int
	tmpCounter int
	funcs      map[string]*ast.FuncDecl
	declaredVars map[string]bool
	inFunc     bool
}

func New() *Generator {
	return &Generator{
		funcs:        make(map[string]*ast.FuncDecl),
		declaredVars: make(map[string]bool),
	}
}

func (g *Generator) emit(s string) {
	for i := 0; i < g.indent; i++ {
		g.out.WriteString("    ")
	}
	g.out.WriteString(s)
	g.out.WriteString("\n")
}

func (g *Generator) emitRaw(s string) {
	g.out.WriteString(s)
}

func (g *Generator) tmpVar() string {
	g.tmpCounter++
	return fmt.Sprintf("_t%d", g.tmpCounter)
}

func (g *Generator) Generate(prog *ast.Program) string {
	// First pass: collect function declarations
	for _, stmt := range prog.Statements {
		if fd, ok := stmt.(*ast.FuncDecl); ok {
			g.funcs[fd.Name] = fd
		}
	}

	// Emit C runtime and header
	g.emitRuntime()

	// Forward declare user functions
	for _, fd := range g.funcs {
		g.emitFuncForwardDecl(fd)
	}
	g.emit("")

	// Emit user function definitions
	for _, stmt := range prog.Statements {
		if fd, ok := stmt.(*ast.FuncDecl); ok {
			g.emitFuncDef(fd)
		}
	}

	// Emit main
	g.emit("int main(int argc, char** argv) {")
	g.indent++
	for _, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.FuncDecl); ok {
			continue
		}
		g.genStmt(stmt)
	}
	g.emit("return 0;")
	g.indent--
	g.emit("}")

	return g.out.String()
}

func (g *Generator) emitRuntime() {
	g.emit(`#include <stdio.h>`)
	g.emit(`#include <stdlib.h>`)
	g.emit(`#include <string.h>`)
	g.emit(`#include <math.h>`)
	g.emit(`#include <time.h>`)
	g.emit(``)
	g.emit(`// Yoft Runtime - Dynamic Value System`)
	g.emit(`typedef enum { VT_NULL, VT_INT, VT_FLOAT, VT_BOOL, VT_STRING, VT_LIST } ValueType;`)
	g.emit(``)
	g.emit(`typedef struct Value {`)
	g.emit(`    ValueType type;`)
	g.emit(`    union {`)
	g.emit(`        long long int_val;`)
	g.emit(`        double float_val;`)
	g.emit(`        int bool_val;`)
	g.emit(`        char* str_val;`)
	g.emit(`        struct { struct Value* items; int len; int cap; } list_val;`)
	g.emit(`    };`)
	g.emit(`} Value;`)
	g.emit(``)
	g.emit(`Value yoft_null() { Value v; v.type = VT_NULL; return v; }`)
	g.emit(`Value yoft_int(long long n) { Value v; v.type = VT_INT; v.int_val = n; return v; }`)
	g.emit(`Value yoft_float(double n) { Value v; v.type = VT_FLOAT; v.float_val = n; return v; }`)
	g.emit(`Value yoft_bool(int b) { Value v; v.type = VT_BOOL; v.bool_val = b; return v; }`)
	g.emit(``)
	g.emit(`Value yoft_string(const char* s) {`)
	g.emit(`    Value v; v.type = VT_STRING;`)
	g.emit(`    v.str_val = malloc(strlen(s) + 1);`)
	g.emit(`    strcpy(v.str_val, s);`)
	g.emit(`    return v;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_list_new(int cap) {`)
	g.emit(`    Value v; v.type = VT_LIST;`)
	g.emit(`    v.list_val.items = (Value*)malloc(sizeof(Value) * (cap > 4 ? cap : 4));`)
	g.emit(`    v.list_val.len = 0;`)
	g.emit(`    v.list_val.cap = cap > 4 ? cap : 4;`)
	g.emit(`    return v;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`void yoft_list_push(Value* list, Value item) {`)
	g.emit(`    if (list->type != VT_LIST) { fprintf(stderr, "TypeError: push on non-list\n"); exit(1); }`)
	g.emit(`    if (list->list_val.len >= list->list_val.cap) {`)
	g.emit(`        list->list_val.cap *= 2;`)
	g.emit(`        list->list_val.items = (Value*)realloc(list->list_val.items, sizeof(Value) * list->list_val.cap);`)
	g.emit(`    }`)
	g.emit(`    list->list_val.items[list->list_val.len++] = item;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_list_pop(Value* list) {`)
	g.emit(`    if (list->type != VT_LIST || list->list_val.len == 0) { fprintf(stderr, "TypeError: pop on empty/non-list\n"); exit(1); }`)
	g.emit(`    return list->list_val.items[--list->list_val.len];`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_index(Value obj, Value idx) {`)
	g.emit(`    if (obj.type == VT_LIST && idx.type == VT_INT) {`)
	g.emit(`        if (idx.int_val < 0 || idx.int_val >= obj.list_val.len) { fprintf(stderr, "IndexError: index out of range\n"); exit(1); }`)
	g.emit(`        return obj.list_val.items[idx.int_val];`)
	g.emit(`    }`)
	g.emit(`    if (obj.type == VT_STRING && idx.type == VT_INT) {`)
	g.emit(`        int slen = strlen(obj.str_val);`)
	g.emit(`        if (idx.int_val < 0 || idx.int_val >= slen) { fprintf(stderr, "IndexError: index out of range\n"); exit(1); }`)
	g.emit(`        char buf[2] = { obj.str_val[idx.int_val], 0 };`)
	g.emit(`        return yoft_string(buf);`)
	g.emit(`    }`)
	g.emit(`    fprintf(stderr, "TypeError: cannot index\n"); exit(1);`)
	g.emit(`    return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`// Convert value to double for arithmetic`)
	g.emit(`double yoft_to_double(Value v) {`)
	g.emit(`    if (v.type == VT_INT) return (double)v.int_val;`)
	g.emit(`    if (v.type == VT_FLOAT) return v.float_val;`)
	g.emit(`    if (v.type == VT_BOOL) return v.bool_val ? 1.0 : 0.0;`)
	g.emit(`    fprintf(stderr, "TypeError: cannot convert to number\n"); exit(1);`)
	g.emit(`    return 0;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`int yoft_is_truthy(Value v) {`)
	g.emit(`    switch (v.type) {`)
	g.emit(`        case VT_NULL: return 0;`)
	g.emit(`        case VT_BOOL: return v.bool_val;`)
	g.emit(`        case VT_INT: return v.int_val != 0;`)
	g.emit(`        case VT_FLOAT: return v.float_val != 0.0;`)
	g.emit(`        case VT_STRING: return strlen(v.str_val) > 0;`)
	g.emit(`        case VT_LIST: return v.list_val.len > 0;`)
	g.emit(`    }`)
	g.emit(`    return 0;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`// String representation`)
	g.emit(`char* yoft_to_str(Value v) {`)
	g.emit(`    char* buf;`)
	g.emit(`    switch (v.type) {`)
	g.emit(`        case VT_NULL: buf = malloc(5); strcpy(buf, "null"); return buf;`)
	g.emit(`        case VT_BOOL: buf = malloc(6); strcpy(buf, v.bool_val ? "true" : "false"); return buf;`)
	g.emit(`        case VT_INT: buf = malloc(32); snprintf(buf, 32, "%lld", v.int_val); return buf;`)
	g.emit(`        case VT_FLOAT: buf = malloc(32); snprintf(buf, 32, "%g", v.float_val); return buf;`)
	g.emit(`        case VT_STRING: buf = malloc(strlen(v.str_val)+1); strcpy(buf, v.str_val); return buf;`)
	g.emit(`        case VT_LIST: {`)
	g.emit(`            int total = 3;`)
	g.emit(`            char** parts = malloc(sizeof(char*) * v.list_val.len);`)
	g.emit(`            for (int i = 0; i < v.list_val.len; i++) {`)
	g.emit(`                parts[i] = yoft_to_str(v.list_val.items[i]);`)
	g.emit(`                total += strlen(parts[i]) + 2;`)
	g.emit(`            }`)
	g.emit(`            buf = malloc(total);`)
	g.emit(`            strcpy(buf, "[");`)
	g.emit(`            for (int i = 0; i < v.list_val.len; i++) {`)
	g.emit(`                if (i > 0) strcat(buf, ", ");`)
	g.emit(`                strcat(buf, parts[i]);`)
	g.emit(`                free(parts[i]);`)
	g.emit(`            }`)
	g.emit(`            strcat(buf, "]");`)
	g.emit(`            free(parts);`)
	g.emit(`            return buf;`)
	g.emit(`        }`)
	g.emit(`    }`)
	g.emit(`    buf = malloc(8); strcpy(buf, "unknown"); return buf;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`void yoft_show(Value v) {`)
	g.emit(`    char* s = yoft_to_str(v);`)
	g.emit(`    printf("%s\n", s);`)
	g.emit(`    free(s);`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`// Arithmetic operations`)
	g.emit(`Value yoft_add(Value a, Value b) {`)
	g.emit(`    if (a.type == VT_STRING || b.type == VT_STRING) {`)
	g.emit(`        char* sa = yoft_to_str(a); char* sb = yoft_to_str(b);`)
	g.emit(`        char* res = malloc(strlen(sa)+strlen(sb)+1);`)
	g.emit(`        strcpy(res, sa); strcat(res, sb);`)
	g.emit(`        free(sa); free(sb);`)
	g.emit(`        Value v; v.type = VT_STRING; v.str_val = res; return v;`)
	g.emit(`    }`)
	g.emit(`    if (a.type == VT_FLOAT || b.type == VT_FLOAT)`)
	g.emit(`        return yoft_float(yoft_to_double(a) + yoft_to_double(b));`)
	g.emit(`    if (a.type == VT_INT && b.type == VT_INT)`)
	g.emit(`        return yoft_int(a.int_val + b.int_val);`)
	g.emit(`    return yoft_float(yoft_to_double(a) + yoft_to_double(b));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_sub(Value a, Value b) {`)
	g.emit(`    if (a.type == VT_FLOAT || b.type == VT_FLOAT)`)
	g.emit(`        return yoft_float(yoft_to_double(a) - yoft_to_double(b));`)
	g.emit(`    if (a.type == VT_INT && b.type == VT_INT)`)
	g.emit(`        return yoft_int(a.int_val - b.int_val);`)
	g.emit(`    return yoft_float(yoft_to_double(a) - yoft_to_double(b));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_mul(Value a, Value b) {`)
	g.emit(`    if (a.type == VT_STRING && b.type == VT_INT) {`)
	g.emit(`        int slen = strlen(a.str_val); long long n = b.int_val;`)
	g.emit(`        char* res = malloc(slen * n + 1); res[0] = 0;`)
	g.emit(`        for (long long i = 0; i < n; i++) strcat(res, a.str_val);`)
	g.emit(`        Value v; v.type = VT_STRING; v.str_val = res; return v;`)
	g.emit(`    }`)
	g.emit(`    if (a.type == VT_FLOAT || b.type == VT_FLOAT)`)
	g.emit(`        return yoft_float(yoft_to_double(a) * yoft_to_double(b));`)
	g.emit(`    if (a.type == VT_INT && b.type == VT_INT)`)
	g.emit(`        return yoft_int(a.int_val * b.int_val);`)
	g.emit(`    return yoft_float(yoft_to_double(a) * yoft_to_double(b));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_div(Value a, Value b) {`)
	g.emit(`    double db = yoft_to_double(b);`)
	g.emit(`    if (db == 0) { fprintf(stderr, "Error: Division by zero\n"); exit(1); }`)
	g.emit(`    if (a.type == VT_INT && b.type == VT_INT)`)
	g.emit(`        return yoft_int(a.int_val / b.int_val);`)
	g.emit(`    return yoft_float(yoft_to_double(a) / db);`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_mod(Value a, Value b) {`)
	g.emit(`    if (a.type == VT_INT && b.type == VT_INT)`)
	g.emit(`        return yoft_int(a.int_val % b.int_val);`)
	g.emit(`    return yoft_float(fmod(yoft_to_double(a), yoft_to_double(b)));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_neg(Value a) {`)
	g.emit(`    if (a.type == VT_INT) return yoft_int(-a.int_val);`)
	g.emit(`    if (a.type == VT_FLOAT) return yoft_float(-a.float_val);`)
	g.emit(`    fprintf(stderr, "TypeError: cannot negate\n"); exit(1);`)
	g.emit(`    return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`// Comparison operations`)
	g.emit(`Value yoft_eq(Value a, Value b) {`)
	g.emit(`    if (a.type == VT_NULL && b.type == VT_NULL) return yoft_bool(1);`)
	g.emit(`    if (a.type == VT_NULL || b.type == VT_NULL) return yoft_bool(0);`)
	g.emit(`    if (a.type == VT_STRING && b.type == VT_STRING) return yoft_bool(strcmp(a.str_val, b.str_val) == 0);`)
	g.emit(`    if (a.type == VT_BOOL && b.type == VT_BOOL) return yoft_bool(a.bool_val == b.bool_val);`)
	g.emit(`    return yoft_bool(yoft_to_double(a) == yoft_to_double(b));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_neq(Value a, Value b) { return yoft_bool(!yoft_is_truthy(yoft_eq(a, b))); }`)
	g.emit(`Value yoft_lt(Value a, Value b) { return yoft_bool(yoft_to_double(a) < yoft_to_double(b)); }`)
	g.emit(`Value yoft_gt(Value a, Value b) { return yoft_bool(yoft_to_double(a) > yoft_to_double(b)); }`)
	g.emit(`Value yoft_lte(Value a, Value b) { return yoft_bool(yoft_to_double(a) <= yoft_to_double(b)); }`)
	g.emit(`Value yoft_gte(Value a, Value b) { return yoft_bool(yoft_to_double(a) >= yoft_to_double(b)); }`)
	g.emit(``)
	g.emit(`// Built-in functions`)
	g.emit(`Value yoft_builtin_len(Value v) {`)
	g.emit(`    if (v.type == VT_STRING) return yoft_int(strlen(v.str_val));`)
	g.emit(`    if (v.type == VT_LIST) return yoft_int(v.list_val.len);`)
	g.emit(`    fprintf(stderr, "TypeError: len() on non-sequence\n"); exit(1);`)
	g.emit(`    return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_type(Value v) {`)
	g.emit(`    switch(v.type) {`)
	g.emit(`        case VT_NULL: return yoft_string("null");`)
	g.emit(`        case VT_INT: return yoft_string("int");`)
	g.emit(`        case VT_FLOAT: return yoft_string("float");`)
	g.emit(`        case VT_BOOL: return yoft_string("bool");`)
	g.emit(`        case VT_STRING: return yoft_string("string");`)
	g.emit(`        case VT_LIST: return yoft_string("list");`)
	g.emit(`    }`)
	g.emit(`    return yoft_string("unknown");`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_int_cast(Value v) {`)
	g.emit(`    if (v.type == VT_INT) return v;`)
	g.emit(`    if (v.type == VT_FLOAT) return yoft_int((long long)v.float_val);`)
	g.emit(`    if (v.type == VT_STRING) return yoft_int(atoll(v.str_val));`)
	g.emit(`    if (v.type == VT_BOOL) return yoft_int(v.bool_val);`)
	g.emit(`    fprintf(stderr, "TypeError: cannot convert to int\n"); exit(1);`)
	g.emit(`    return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_float_cast(Value v) {`)
	g.emit(`    if (v.type == VT_FLOAT) return v;`)
	g.emit(`    if (v.type == VT_INT) return yoft_float((double)v.int_val);`)
	g.emit(`    if (v.type == VT_STRING) return yoft_float(atof(v.str_val));`)
	g.emit(`    if (v.type == VT_BOOL) return yoft_float(v.bool_val ? 1.0 : 0.0);`)
	g.emit(`    fprintf(stderr, "TypeError: cannot convert to float\n"); exit(1);`)
	g.emit(`    return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_str_cast(Value v) { char* s = yoft_to_str(v); Value r; r.type = VT_STRING; r.str_val = s; return r; }`)
	g.emit(``)
	g.emit(`Value yoft_builtin_input(Value prompt) {`)
	g.emit(`    if (prompt.type == VT_STRING) printf("%s", prompt.str_val);`)
	g.emit(`    char buf[4096]; if (fgets(buf, 4096, stdin)) { buf[strcspn(buf, "\n")] = 0; return yoft_string(buf); }`)
	g.emit(`    return yoft_string("");`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_abs(Value v) {`)
	g.emit(`    if (v.type == VT_INT) return yoft_int(v.int_val < 0 ? -v.int_val : v.int_val);`)
	g.emit(`    if (v.type == VT_FLOAT) return yoft_float(v.float_val < 0 ? -v.float_val : v.float_val);`)
	g.emit(`    fprintf(stderr, "TypeError: abs() requires number\n"); exit(1); return yoft_null();`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_rand(Value a, Value b) {`)
	g.emit(`    static int seeded = 0; if (!seeded) { srand(time(NULL)); seeded = 1; }`)
	g.emit(`    int lo = (int)a.int_val, hi = (int)b.int_val;`)
	g.emit(`    return yoft_int(lo + rand() % (hi - lo + 1));`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_range(Value start, Value end) {`)
	g.emit(`    Value list = yoft_list_new((int)(end.int_val - start.int_val));`)
	g.emit(`    for (long long i = start.int_val; i < end.int_val; i++) yoft_list_push(&list, yoft_int(i));`)
	g.emit(`    return list;`)
	g.emit(`}`)
	g.emit(``)
	g.emit(`Value yoft_builtin_push(Value* list, Value item) { yoft_list_push(list, item); return *list; }`)
	g.emit(`Value yoft_builtin_pop(Value* list) { return yoft_list_pop(list); }`)
	g.emit(``)
	g.emit(`Value yoft_builtin_min(Value a, Value b) { return yoft_to_double(a) < yoft_to_double(b) ? a : b; }`)
	g.emit(`Value yoft_builtin_max(Value a, Value b) { return yoft_to_double(a) > yoft_to_double(b) ? a : b; }`)
	g.emit(`Value yoft_builtin_round(Value v) { return yoft_int((long long)(yoft_to_double(v) + 0.5)); }`)
	g.emit(``)
	g.emit(`// String methods`)
	g.emit(`Value yoft_str_length(Value s) { return yoft_int(strlen(s.str_val)); }`)
	g.emit(`Value yoft_str_upper(Value s) {`)
	g.emit(`    char* r = malloc(strlen(s.str_val)+1); int i;`)
	g.emit(`    for(i=0;s.str_val[i];i++) r[i]=s.str_val[i]>='a'&&s.str_val[i]<='z'?s.str_val[i]-32:s.str_val[i];`)
	g.emit(`    r[i]=0; Value v; v.type=VT_STRING; v.str_val=r; return v;`)
	g.emit(`}`)
	g.emit(`Value yoft_str_lower(Value s) {`)
	g.emit(`    char* r = malloc(strlen(s.str_val)+1); int i;`)
	g.emit(`    for(i=0;s.str_val[i];i++) r[i]=s.str_val[i]>='A'&&s.str_val[i]<='Z'?s.str_val[i]+32:s.str_val[i];`)
	g.emit(`    r[i]=0; Value v; v.type=VT_STRING; v.str_val=r; return v;`)
	g.emit(`}`)
	g.emit(`Value yoft_str_contains(Value s, Value sub) { return yoft_bool(strstr(s.str_val, sub.str_val) != NULL); }`)
	g.emit(`Value yoft_list_length(Value l) { return yoft_int(l.list_val.len); }`)
	g.emit(`Value yoft_list_contains(Value l, Value item) {`)
	g.emit(`    for(int i=0;i<l.list_val.len;i++) if(yoft_is_truthy(yoft_eq(l.list_val.items[i],item))) return yoft_bool(1);`)
	g.emit(`    return yoft_bool(0);`)
	g.emit(`}`)
	g.emit(`Value yoft_list_reverse(Value* l) {`)
	g.emit(`    for(int i=0,j=l->list_val.len-1;i<j;i++,j--) { Value t=l->list_val.items[i]; l->list_val.items[i]=l->list_val.items[j]; l->list_val.items[j]=t; }`)
	g.emit(`    return *l;`)
	g.emit(`}`)
	g.emit(`Value yoft_list_join(Value l, Value sep) {`)
	g.emit(`    int total = 1;`)
	g.emit(`    char** parts = malloc(sizeof(char*)*l.list_val.len);`)
	g.emit(`    for(int i=0;i<l.list_val.len;i++) { parts[i]=yoft_to_str(l.list_val.items[i]); total+=strlen(parts[i])+strlen(sep.str_val); }`)
	g.emit(`    char* buf=malloc(total); buf[0]=0;`)
	g.emit(`    for(int i=0;i<l.list_val.len;i++) { if(i>0) strcat(buf,sep.str_val); strcat(buf,parts[i]); free(parts[i]); }`)
	g.emit(`    free(parts); Value v; v.type=VT_STRING; v.str_val=buf; return v;`)
	g.emit(`}`)
	g.emit(``)
}

func (g *Generator) emitFuncForwardDecl(fd *ast.FuncDecl) {
	params := make([]string, len(fd.Params))
	for i := range fd.Params {
		params[i] = "Value"
	}
	g.emit(fmt.Sprintf("Value yoft_func_%s(%s);", fd.Name, strings.Join(params, ", ")))
}

func (g *Generator) emitFuncDef(fd *ast.FuncDecl) {
	params := make([]string, len(fd.Params))
	for i, p := range fd.Params {
		params[i] = fmt.Sprintf("Value %s", g.safeVar(p))
	}
	g.emit(fmt.Sprintf("Value yoft_func_%s(%s) {", fd.Name, strings.Join(params, ", ")))
	g.indent++
	prevInFunc := g.inFunc
	prevDeclared := g.declaredVars
	g.inFunc = true
	g.declaredVars = make(map[string]bool)
	for _, p := range fd.Params {
		g.declaredVars[p] = true
	}
	for _, stmt := range fd.Body {
		g.genStmt(stmt)
	}
	g.emit("return yoft_null();")
	g.inFunc = prevInFunc
	g.declaredVars = prevDeclared
	g.indent--
	g.emit("}")
	g.emit("")
}

func (g *Generator) safeVar(name string) string {
	return "v_" + name
}

func (g *Generator) genStmt(node ast.Node) {
	switch n := node.(type) {
	case *ast.VarDecl:
		val := g.genExpr(n.Value)
		if g.declaredVars[n.Name] {
			g.emit(fmt.Sprintf("%s = %s;", g.safeVar(n.Name), val))
		} else {
			g.emit(fmt.Sprintf("Value %s = %s;", g.safeVar(n.Name), val))
			g.declaredVars[n.Name] = true
		}

	case *ast.VarReassign:
		val := g.genExpr(n.Value)
		g.emit(fmt.Sprintf("%s = %s;", g.safeVar(n.Name), val))

	case *ast.ShowStmt:
		val := g.genExpr(n.Value)
		g.emit(fmt.Sprintf("yoft_show(%s);", val))

	case *ast.IfStmt:
		cond := g.genExpr(n.Condition)
		g.emit(fmt.Sprintf("if (yoft_is_truthy(%s)) {", cond))
		g.indent++
		for _, s := range n.Body {
			g.genStmt(s)
		}
		g.indent--
		if n.ElseBody != nil {
			g.emit("} else {")
			g.indent++
			for _, s := range n.ElseBody {
				g.genStmt(s)
			}
			g.indent--
		}
		g.emit("}")

	case *ast.WhileStmt:
		cond := g.genExpr(n.Condition)
		g.emit(fmt.Sprintf("while (yoft_is_truthy(%s)) {", cond))
		g.indent++
		for _, s := range n.Body {
			g.genStmt(s)
		}
		// Re-evaluate condition variable (for while loops that modify variables)
		g.emit(fmt.Sprintf("// re-check: %s", cond))
		g.indent--
		g.emit("}")

	case *ast.ForStmt:
		iterVar := g.tmpVar()
		idxVar := g.tmpVar()
		iterExpr := g.genExpr(n.Iterable)
		g.emit(fmt.Sprintf("Value %s = %s;", iterVar, iterExpr))
		g.emit(fmt.Sprintf("for (int %s = 0; %s < %s.list_val.len; %s++) {", idxVar, idxVar, iterVar, idxVar))
		g.indent++
		if !g.declaredVars[n.VarName] {
			g.emit(fmt.Sprintf("Value %s = %s.list_val.items[%s];", g.safeVar(n.VarName), iterVar, idxVar))
			g.declaredVars[n.VarName] = true
		} else {
			g.emit(fmt.Sprintf("%s = %s.list_val.items[%s];", g.safeVar(n.VarName), iterVar, idxVar))
		}
		for _, s := range n.Body {
			g.genStmt(s)
		}
		g.indent--
		g.emit("}")

	case *ast.ReturnStmt:
		if n.Value != nil {
			val := g.genExpr(n.Value)
			g.emit(fmt.Sprintf("return %s;", val))
		} else {
			g.emit("return yoft_null();")
		}

	case *ast.FuncDecl:
		// Already handled at top level
		return

	default:
		// Expression statement
		expr := g.genExpr(node)
		g.emit(fmt.Sprintf("%s;", expr))
	}
}

func (g *Generator) genExpr(node ast.Node) string {
	switch n := node.(type) {
	case *ast.NumberLiteral:
		if n.IsFloat {
			return fmt.Sprintf("yoft_float(%s)", n.Value)
		}
		return fmt.Sprintf("yoft_int(%s)", n.Value)

	case *ast.StringLiteral:
		escaped := strings.ReplaceAll(n.Value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "\n", "\\n")
		escaped = strings.ReplaceAll(escaped, "\t", "\\t")
		return fmt.Sprintf("yoft_string(\"%s\")", escaped)

	case *ast.BoolLiteral:
		if n.Value {
			return "yoft_bool(1)"
		}
		return "yoft_bool(0)"

	case *ast.NullLiteral:
		return "yoft_null()"

	case *ast.Identifier:
		return g.safeVar(n.Name)

	case *ast.ListLiteral:
		tmp := g.tmpVar()
		g.emit(fmt.Sprintf("Value %s = yoft_list_new(%d);", tmp, len(n.Elements)))
		for _, el := range n.Elements {
			val := g.genExpr(el)
			g.emit(fmt.Sprintf("yoft_list_push(&%s, %s);", tmp, val))
		}
		return tmp

	case *ast.BinaryOp:
		left := g.genExpr(n.Left)
		right := g.genExpr(n.Right)
		switch n.Op {
		case "+":
			return fmt.Sprintf("yoft_add(%s, %s)", left, right)
		case "-":
			return fmt.Sprintf("yoft_sub(%s, %s)", left, right)
		case "*":
			return fmt.Sprintf("yoft_mul(%s, %s)", left, right)
		case "/":
			return fmt.Sprintf("yoft_div(%s, %s)", left, right)
		case "%":
			return fmt.Sprintf("yoft_mod(%s, %s)", left, right)
		case "==":
			return fmt.Sprintf("yoft_eq(%s, %s)", left, right)
		case "!=":
			return fmt.Sprintf("yoft_neq(%s, %s)", left, right)
		case "<":
			return fmt.Sprintf("yoft_lt(%s, %s)", left, right)
		case ">":
			return fmt.Sprintf("yoft_gt(%s, %s)", left, right)
		case "<=":
			return fmt.Sprintf("yoft_lte(%s, %s)", left, right)
		case ">=":
			return fmt.Sprintf("yoft_gte(%s, %s)", left, right)
		case "and":
			return fmt.Sprintf("(yoft_is_truthy(%s) ? %s : %s)", left, right, left)
		case "or":
			return fmt.Sprintf("(yoft_is_truthy(%s) ? %s : %s)", left, left, right)
		}

	case *ast.UnaryOp:
		operand := g.genExpr(n.Operand)
		if n.Op == "-" {
			return fmt.Sprintf("yoft_neg(%s)", operand)
		}
		if n.Op == "not" {
			return fmt.Sprintf("yoft_bool(!yoft_is_truthy(%s))", operand)
		}

	case *ast.FuncCall:
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			args[i] = g.genExpr(a)
		}
		argStr := strings.Join(args, ", ")

		// Built-in functions
		switch n.Name {
		case "show":
			if len(args) > 0 {
				return fmt.Sprintf("(yoft_show(%s), yoft_null())", args[0])
			}
			return "(printf(\"\\n\"), yoft_null())"
		case "len":
			return fmt.Sprintf("yoft_builtin_len(%s)", argStr)
		case "type":
			return fmt.Sprintf("yoft_builtin_type(%s)", argStr)
		case "int":
			return fmt.Sprintf("yoft_builtin_int_cast(%s)", argStr)
		case "float":
			return fmt.Sprintf("yoft_builtin_float_cast(%s)", argStr)
		case "str":
			return fmt.Sprintf("yoft_builtin_str_cast(%s)", argStr)
		case "input":
			if len(args) == 0 {
				return "yoft_builtin_input(yoft_string(\"\"))"
			}
			return fmt.Sprintf("yoft_builtin_input(%s)", argStr)
		case "abs":
			return fmt.Sprintf("yoft_builtin_abs(%s)", argStr)
		case "rand":
			return fmt.Sprintf("yoft_builtin_rand(%s)", argStr)
		case "range":
			if len(args) == 1 {
				return fmt.Sprintf("yoft_builtin_range(yoft_int(0), %s)", args[0])
			}
			return fmt.Sprintf("yoft_builtin_range(%s)", argStr)
		case "push":
			return fmt.Sprintf("yoft_builtin_push(&%s, %s)", args[0], args[1])
		case "pop":
			return fmt.Sprintf("yoft_builtin_pop(&%s)", args[0])
		case "min":
			return fmt.Sprintf("yoft_builtin_min(%s)", argStr)
		case "max":
			return fmt.Sprintf("yoft_builtin_max(%s)", argStr)
		case "round":
			return fmt.Sprintf("yoft_builtin_round(%s)", argStr)
		}

		// User function
		return fmt.Sprintf("yoft_func_%s(%s)", n.Name, argStr)

	case *ast.IndexAccess:
		obj := g.genExpr(n.Object)
		idx := g.genExpr(n.Index)
		return fmt.Sprintf("yoft_index(%s, %s)", obj, idx)

	case *ast.MethodCall:
		obj := g.genExpr(n.Object)
		args := make([]string, len(n.Args))
		for i, a := range n.Args {
			args[i] = g.genExpr(a)
		}
		switch n.Method {
		case "length":
			return fmt.Sprintf("yoft_builtin_len(%s)", obj)
		case "upper":
			return fmt.Sprintf("yoft_str_upper(%s)", obj)
		case "lower":
			return fmt.Sprintf("yoft_str_lower(%s)", obj)
		case "contains":
			return fmt.Sprintf("yoft_str_contains(%s, %s)", obj, args[0])
		case "push":
			return fmt.Sprintf("yoft_builtin_push(&%s, %s)", obj, args[0])
		case "pop":
			return fmt.Sprintf("yoft_builtin_pop(&%s)", obj)
		case "join":
			return fmt.Sprintf("yoft_list_join(%s, %s)", obj, args[0])
		case "reverse":
			return fmt.Sprintf("yoft_list_reverse(&%s)", obj)
		}
		return fmt.Sprintf("/* unknown method %s */ yoft_null()", n.Method)
	}
	return "yoft_null()"
}
